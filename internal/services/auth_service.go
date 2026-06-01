package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/parxyws/cozybox/internal/core"
	"github.com/parxyws/cozybox/internal/dto"
	"github.com/parxyws/cozybox/internal/models"
	"github.com/parxyws/cozybox/pkg/database/aws"
	helper2 "github.com/parxyws/cozybox/pkg/helper"
	"github.com/parxyws/cozybox/pkg/mail"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db        *gorm.DB
	mailer    *mail.Mailer
	authRds   *redis.Client
	s3service *aws.S3Service
	jwt       helper2.TokenGenerator
	addr      string
}

const (
	OTP_LENGTH = 6
	OTP_EXPIRY = 2 * time.Minute
)

func NewUserService(db *gorm.DB, mailer *mail.Mailer, authRds *redis.Client, jwt helper2.TokenGenerator, addr string, s3service *aws.S3Service) *AuthService {
	return &AuthService{db: db, mailer: mailer, authRds: authRds, jwt: jwt, addr: addr, s3service: s3service}
}

func (u *AuthService) Register(ctx context.Context, req *dto.RegisterUserRequest) (*dto.RegisterResponse, error) {
	currentTime := time.Now()
	identifier := base64.StdEncoding.EncodeToString([]byte(req.Email))
	referenceId := fmt.Sprintf("%s-%s", identifier, ulid.Make().String()) // generate random string for email verification
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to generate password")
	}

	request := &models.User{
		Id:                  ulid.Make().String(),
		Name:                req.Name,
		Username:            req.Username,
		Email:               req.Email,
		Password:            string(hashedPassword),
		IsVerified:          false,
		ForcePasswordChange: false,
		OnboardingCompleted: false,
		CreatedAt:           currentTime,
		UpdatedAt:           currentTime,
	}

	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	res := tx.Where("email = ? AND deleted_at IS NULL", req.Email).FirstOrCreate(request)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to register user: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return nil, errors.New("user already exists")
	}

	// Create default tenant for the new user
	tenantId := ulid.Make().String()
	currentSlug := generateSlug(req.TenantName, tenantId)
	tenant := &models.Tenant{
		Id:        tenantId,
		Name:      req.TenantName,
		Slug:      currentSlug,
		Status:    models.TenantStatusActive,
		OwnerId:   request.Id,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	if err := tx.Create(tenant).Error; err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Add user as owner member of the tenant
	member := &models.TenantMember{
		Id:       ulid.Make().String(),
		TenantId: tenant.Id,
		UserId:   request.Id,
		Role:     models.TenantRoleOwner,
		JoinedAt: currentTime,
	}

	if err := tx.Create(member).Error; err != nil {
		return nil, fmt.Errorf("failed to create tenant member: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	marshaledData, err := json.Marshal(dto.UserResponse{
		Id:       request.Id,
		Name:     request.Name,
		Username: request.Username,
		Email:    request.Email,
	})
	// TODO: provide more generic error message
	if err != nil {
		return nil, errors.New("failed to marshal user data")
	}

	otp, err := helper2.GenerateRandomInteger(OTP_LENGTH)
	if err != nil {
		return nil, errors.New("failed to generate otp")
	}

	// Store OTP and reference data in Redis using Pipeline (single round-trip)
	pipe := u.authRds.Pipeline()
	pipe.Set(ctx, fmt.Sprintf("ref-%s", referenceId), marshaledData, OTP_EXPIRY)
	pipe.Set(ctx, fmt.Sprintf("otp-%s", referenceId), otp, OTP_EXPIRY)
	if _, err = pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to store registration data: %w", err)
	}

	go func() {
		otpData := &mail.OTPData{
			Name: request.Name,
			OTP:  otp,
		}
		if err := u.mailer.SendOTP(request.Email, "Verify Your Email", otpData); err != nil {
			fmt.Printf("failed to send otp to %s: %v\n", request.Email, err)
		}
	}()

	return &dto.RegisterResponse{
		ReferenceId: referenceId,
	}, nil
}

func (u *AuthService) VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) (*dto.VerifyEmailResponse, error) {
	trimmedReferenceID := strings.TrimSpace(req.ReferenceId)

	if err := u.verifyOTP(ctx, trimmedReferenceID, req.Otp); err != nil {
		return nil, err
	}

	data, err := u.authRds.Get(ctx, fmt.Sprintf("ref-%s", trimmedReferenceID)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, errors.New("reference id not found")
	} else if err != nil {
		return nil, errors.New("failed to get reference mappings")
	}

	var user dto.UserResponse
	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		return nil, errors.New("failed to unmarshal user data")
	}

	var userModel models.User
	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()
	res := tx.Where("id = ? AND deleted_at IS NULL", user.Id).First(&userModel)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to get user: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}

	userModel.IsVerified = true
	userModel.UpdatedAt = time.Now()
	res = tx.Updates(&userModel)
	if res.Error != nil {
		return nil, errors.New("failed to verify email")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &dto.VerifyEmailResponse{
		ReferenceId: trimmedReferenceID,
	}, nil
}

func (u *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.UserAuthenticateResponse, error) {
	var userModel models.User

	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	res := tx.Where("email = ? AND deleted_at IS NULL", req.Email).First(&userModel)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, fmt.Errorf("failed to get user: %w", res.Error)
	}

	err := bcrypt.CompareHashAndPassword([]byte(userModel.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !userModel.IsVerified {
		return nil, errors.New("email is not verified")
	}

	sessionId := ulid.Make().String()

	// Resolve the user's active tenant
	var member models.TenantMember
	if err := u.db.WithContext(ctx).
		Where("user_id = ?", userModel.Id).
		Preload("Tenant").
		First(&member).Error; err != nil {
		return nil, errors.New("no tenant found for user")
	}

	accessToken, refreshToken, hashedRefreshToken, err := u.generateAccessToken(userModel.Id, member.TenantId, sessionId)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	session := models.UserSession{
		SessionID:    sessionId,
		UserID:       userModel.Id,
		TenantID:     member.TenantId,
		RefreshToken: string(hashedRefreshToken),
		ClientIP:     "",                                 // TODO: Get from context/request
		UserAgent:    "",                                 // TODO: Get from context/request
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 Days
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, errors.New("failed to marshal session")
	}

	err = u.authRds.Set(ctx, fmt.Sprintf("session:%s", sessionId), sessionJSON, 7*24*time.Hour).Err()
	if err != nil {
		return nil, errors.New("failed to store session in redis")
	}

	userModel.LastLogin = time.Now()
	if err := tx.Save(&userModel).Error; err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &dto.UserAuthenticateResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		User: dto.UserResponse{
			Id:                  userModel.Id,
			Name:                userModel.Name,
			Username:            userModel.Username,
			Email:               userModel.Email,
			ForcePasswordChange: userModel.ForcePasswordChange,
			OnboardingCompleted: userModel.OnboardingCompleted,
		},
		Tenant: dto.TenantResponse{
			Id:   member.TenantId,
			Name: member.Tenant.Name,
			Slug: member.Tenant.Slug,
			Role: string(member.Role),
		},
	}, nil
}

func (u *AuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.UserAuthenticateResponse, error) {
	sessionJSON, err := u.authRds.Get(ctx, fmt.Sprintf("session:%s", req.SessionID)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, errors.New("invalid or expired session")
	} else if err != nil {
		return nil, errors.New("failed to retrieve session")
	}

	var session models.UserSession
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return nil, errors.New("failed to unmarshal session")
	}

	// Verify the refresh token matches
	err = bcrypt.CompareHashAndPassword([]byte(session.RefreshToken), []byte(req.RefreshToken))
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if time.Now().After(session.ExpiresAt) {
		_ = u.authRds.Del(ctx, fmt.Sprintf("session:%s", req.SessionID))
		return nil, errors.New("refresh token expired")
	}

	// Get user details
	var userModel models.User
	res := u.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", session.UserID).First(&userModel)
	if res.Error != nil {
		return nil, errors.New("user not found")
	}

	// ---- TOKEN ROTATION ----

	// Delete old session
	_ = u.authRds.Del(ctx, fmt.Sprintf("session:%s", req.SessionID))

	// Generate new session & tokens
	newSessionID := ulid.Make().String()

	newAccessToken, newRefreshToken, hashedRefreshToken, err := u.generateAccessToken(userModel.Id, session.TenantID, newSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newSession := models.UserSession{
		SessionID:    newSessionID,
		UserID:       userModel.Id,
		TenantID:     session.TenantID,
		RefreshToken: string(hashedRefreshToken),
		ClientIP:     session.ClientIP, // preserve original client info
		UserAgent:    session.UserAgent,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	newSessionJSON, err := json.Marshal(newSession)
	if err != nil {
		return nil, errors.New("failed to marshal new session")
	}

	err = u.authRds.Set(ctx, fmt.Sprintf("session:%s", newSessionID), newSessionJSON, 7*24*time.Hour).Err()
	if err != nil {
		return nil, errors.New("failed to store new session in redis")
	}

	return &dto.UserAuthenticateResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		User: dto.UserResponse{
			Id:                  userModel.Id,
			Name:                userModel.Name,
			Username:            userModel.Username,
			Email:               userModel.Email,
			ForcePasswordChange: userModel.ForcePasswordChange,
			OnboardingCompleted: userModel.OnboardingCompleted,
		},
		Tenant: dto.TenantResponse{
			Id: session.TenantID,
		},
	}, nil
}

func (u *AuthService) Logout(ctx context.Context, sessionID string) error {
	err := u.authRds.Del(ctx, fmt.Sprintf("session:%s", sessionID)).Err()
	if err != nil {
		return errors.New("failed to delete session")
	}
	return nil
}

func (u *AuthService) CompleteOnboarding(ctx context.Context, req *dto.OnboardingRequest, image *aws.UploadInput) error {
	id, ok := ctx.Value(core.UserID).(string)
	tenantId, ok := ctx.Value(core.TenantID).(string)
	if !ok || id == "" {
		return errors.New("unauthorized: user context missing")
	}

	var (
		tenantData models.Tenant
		userData   models.User
		tenantErr  error
		userErr    error
	)

	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		tenantErr = tx.First(&tenantData, "id = ?", tenantId).Error
	}()

	go func() {
		defer wg.Done()
		userErr = tx.First(&userData, "id = ?", id).Error
	}()

	wg.Wait()

	if tenantErr != nil || userErr != nil {
		return errors.New("invalid or expired tenant/user")
	}

	imgResponse, err := u.s3service.PutObject(ctx, image)
	if err != nil {
		return errors.New("failed to upload image")
	}

	organization := &models.Organization{
		Id:           ulid.Make().String(),
		TenantId:     tenantId,
		Name:         tenantData.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		State:        req.State,
		PostalCode:   req.PostalCode,
		Country:      req.Country,
		TaxId:        req.TaxId,
		LogoS3Key:    imgResponse.Key,
	}

	if err := tx.Save(organization).Error; err != nil {
		return errors.New("failed to save data")
	}

	userData.OnboardingCompleted = true
	userData.UpdatedAt = time.Now()

	if err := tx.Updates(&userData).Error; err != nil {
		return errors.New("failed to complete onboarding")
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("failed to commit onboarding")
	}

	return nil
}

// ------------- User Service Tools --------------------

func (u *AuthService) verifyOTP(ctx context.Context, reference string, otp string) error {
	trimmedReference := strings.TrimSpace(reference)
	trimmedOTP := strings.TrimSpace(otp)

	otp, err := u.authRds.Get(ctx, fmt.Sprintf("otp-%s", trimmedReference)).Result()
	if errors.Is(err, redis.Nil) {
		return errors.New("otp not found")
	} else if err != nil {
		return errors.New("failed to check otp mapping")
	}

	if otp != trimmedOTP {
		return errors.New("otp not match")
	}

	if err := u.authRds.Del(ctx, fmt.Sprintf("otp-%s", trimmedReference)).Err(); err != nil {
		return errors.New("failed to delete otp")
	}

	return nil
}

// generateSlug converts a name into a URL-safe slug using the tenant ID for uniqueness.
// Example: "John Doe" with id "01JFXYZ123ABC" → "john-doe-01jfxyz1"
func generateSlug(name string, id string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	// Replace non-alphanumeric characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "workspace"
	}
	// Use the first 8 characters of the tenant ID as suffix
	suffix := strings.ToLower(id[:8])
	return fmt.Sprintf("%s-%s", slug, suffix)
}

func (u *AuthService) generateAccessToken(userId, tenantId, sessionId string) (string, string, []byte, error) {
	accessToken, err := u.jwt.CreateAccessToken(userId, tenantId, sessionId, 15*time.Minute)
	if err != nil {
		return "", "", nil, err
	}

	refreshToken, err := u.jwt.GenerateRefreshToken()
	if err != nil {
		return "", "", nil, err
	}

	// Hash refresh token before saving to database
	hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, hashedRefreshToken, nil
}

func (u *AuthService) ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error {
	identifier := base64.StdEncoding.EncodeToString([]byte(req.Email))
	referenceId := fmt.Sprintf("%s-%s", identifier, ulid.Make().String())

	otp, err := helper2.GenerateRandomInteger(OTP_LENGTH)

	pipe := u.authRds.Pipeline()
	pipe.Set(ctx, fmt.Sprintf("otp-%s", referenceId), otp, OTP_EXPIRY)
	if _, err = pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to store registration data: %w", err)
	}

	go func() {
		data := &mail.ResetPasswordData{
			Name: req.Email,
			Url:  fmt.Sprintf("%s/reset-password?ref=%s&token=%s", u.addr, referenceId, otp),
		}
		if err := u.mailer.SendResetPassword(req.Email, "Reset Password", data); err != nil {
			log.Printf("failed to send reset password: %v", err)
		}
	}()

	return nil
}

func (u *AuthService) ResetPassword(ctx context.Context, request *dto.ResetPasswordRequest) error {
	currentTime := time.Now()
	trimmedReferenceID := strings.TrimSpace(request.ReferenceId)
	err := u.verifyOTP(ctx, trimmedReferenceID, request.Token)
	if err != nil {
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)

	if errors.Is(err, redis.Nil) {
		return errors.New("reset link is invalid or has expired")
	}
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	parts := strings.SplitN(request.ReferenceId, "-", 2)
	if len(parts) < 2 {
		return errors.New("invalid reference ID")
	}

	email, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("failed to decode email: %w", err)
	}

	var userModel models.User
	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	res := tx.Where("email = ? AND deleted_at IS NULL", email).First(&userModel)
	if res.Error != nil {
		return fmt.Errorf("failed to get user: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return errors.New("user not found")
	}

	userModel.Password = string(hashedPassword)
	userModel.UpdatedAt = currentTime

	if err := tx.Updates(&userModel).Error; err != nil {
		return errors.New("failed to update password")
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New("failed to commit")
	}

	return nil
}
