package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/pkg/jwtutil"
	"github.com/parxyws/cozybox/internal/pkg/mail"
	"github.com/parxyws/cozybox/internal/pkg/randutil"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Insert(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
}

type TenantRepository interface {
	Insert(ctx context.Context, tenant *domain.Tenant) error
	Update(ctx context.Context, tenant *domain.Tenant) error
	GetByID(ctx context.Context, id string) (*domain.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
}

type TenantMemberRepository interface {
	Insert(ctx context.Context, member *domain.TenantMember) error
	GetByUserID(ctx context.Context, userID string) (*domain.TenantMember, error)
}

type OrganizationRepository interface {
	Insert(ctx context.Context, org *domain.Organization) error
}

type SessionStore interface {
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	SetMultiple(ctx context.Context, pairs map[string]string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, keys ...string) error
}

type Mailer interface {
	SendOTP(to string, data mail.OTPData) error
	SendResetPassword(to string, data mail.ResetPasswordData) error
}

type FileStorage interface {
	PutObject(ctx context.Context, input domain.UploadInput) (key string, err error)
}

type Service struct {
	userRepo       UserRepository
	tenantRepo     TenantRepository
	memberRepo     TenantMemberRepository
	orgRepo        OrganizationRepository
	sessionStore   SessionStore
	tokenGenerator jwtutil.TokenGenerator
	mailer         Mailer
	fileStorage    FileStorage
	addr           string
	wg             sync.WaitGroup
}

const (
	otpLength = 6
	otpExpiry = 2 * time.Minute
)

func NewAuthService(
	userRepo UserRepository,
	tenantRepo TenantRepository,
	memberRepo TenantMemberRepository,
	orgRepo OrganizationRepository,
	sessionStore SessionStore,
	tokenGenerator jwtutil.TokenGenerator,
	mailer Mailer,
	fileStorage FileStorage,
	addr string,
) *Service {
	return &Service{
		userRepo:       userRepo,
		tenantRepo:     tenantRepo,
		memberRepo:     memberRepo,
		orgRepo:        orgRepo,
		sessionStore:   sessionStore,
		tokenGenerator: tokenGenerator,
		mailer:         mailer,
		fileStorage:    fileStorage,
		addr:           addr,
	}
}

func (s *Service) Register(ctx context.Context, req *RegisterUserRequest) (*RegisterResponse, error) {
	currentTime := time.Now()
	identifier := base64.StdEncoding.EncodeToString([]byte(req.Email))
	referenceId := fmt.Sprintf("%s-%s", identifier, ulid.Make().String())
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	user := &domain.User{
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

	if err := s.userRepo.Insert(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	tenantId := ulid.Make().String()
	slug := generateSlug(req.TenantName, tenantId)
	tenant := &domain.Tenant{
		Id:        tenantId,
		Name:      req.TenantName,
		Slug:      slug,
		Status:    domain.TenantStatusActive,
		OwnerId:   user.Id,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	if err := s.tenantRepo.Insert(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	member := &domain.TenantMember{
		Id:       ulid.Make().String(),
		TenantId: tenant.Id,
		UserId:   user.Id,
		Role:     domain.TenantRoleOwner,
		JoinedAt: currentTime,
	}

	if err := s.memberRepo.Insert(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to create tenant member: %w", err)
	}

	marshaledData, err := json.Marshal(UserResponse{
		Id:       user.Id,
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user data: %w", err)
	}

	otp, err := randutil.GenerateRandomInteger(otpLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate otp: %w", err)
	}

	if err := s.sessionStore.SetMultiple(ctx, map[string]string{
		fmt.Sprintf("ref-%s", referenceId): string(marshaledData),
		fmt.Sprintf("otp-%s", referenceId): otp,
	}, otpExpiry); err != nil {
		return nil, fmt.Errorf("failed to store registration data: %w", err)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.mailer.SendOTP(req.Email, mail.OTPData{Name: req.Name, OTP: otp}); err != nil {
			log.Printf("failed to send otp to %s: %v", req.Email, err)
		}
	}()

	return &RegisterResponse{ReferenceId: referenceId}, nil
}

func (s *Service) Shutdown() error {
	s.wg.Wait()
	return nil
}

func (s *Service) VerifyEmail(ctx context.Context, req *VerifyEmailRequest) (*VerifyEmailResponse, error) {
	trimmedRef := strings.TrimSpace(req.ReferenceId)

	if err := s.verifyOTP(ctx, trimmedRef, req.Otp); err != nil {
		return nil, err
	}

	data, err := s.sessionStore.Get(ctx, fmt.Sprintf("ref-%s", trimmedRef))
	if err != nil {
		return nil, domain.ErrNotFound
	}

	var user UserResponse
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	existingUser, err := s.userRepo.GetByID(ctx, user.Id)
	if err != nil {
		return nil, err
	}

	existingUser.IsVerified = true
	existingUser.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		return nil, fmt.Errorf("failed to verify email: %w", err)
	}

	return &VerifyEmailResponse{ReferenceId: trimmedRef}, nil
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (*UserAuthenticateResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredential
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidCredential
	}

	if !user.IsVerified {
		return nil, domain.ErrEmailNotVerified
	}

	member, err := s.memberRepo.GetByUserID(ctx, user.Id)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	sessionId := ulid.Make().String()
	accessToken, refreshToken, hashedToken, err := s.generateTokens(user.Id, member.TenantId, sessionId)
	if err != nil {
		return nil, err
	}

	session := domain.UserSession{
		SessionID:    sessionId,
		UserID:       user.Id,
		TenantID:     member.TenantId,
		RefreshToken: string(hashedToken),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := s.sessionStore.Set(ctx, fmt.Sprintf("session:%s", sessionId), string(sessionJSON), 7*24*time.Hour); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	user.LastLogin = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	tenant, err := s.tenantRepo.GetByID(ctx, member.TenantId)
	if err != nil {
		return nil, err
	}

	return &UserAuthenticateResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		User: UserResponse{
			Id:                  user.Id,
			Name:                user.Name,
			Username:            user.Username,
			Email:               user.Email,
			ForcePasswordChange: user.ForcePasswordChange,
			OnboardingCompleted: user.OnboardingCompleted,
		},
		Tenant: TenantResponse{
			Id:   tenant.Id,
			Name: tenant.Name,
			Slug: tenant.Slug,
			Role: string(member.Role),
		},
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*UserAuthenticateResponse, error) {
	sessionJSON, err := s.sessionStore.Get(ctx, fmt.Sprintf("session:%s", req.SessionID))
	if err != nil {
		return nil, domain.ErrSessionExpired
	}

	var session domain.UserSession
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(session.RefreshToken), []byte(req.RefreshToken)); err != nil {
		return nil, domain.ErrUnauthorized
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.sessionStore.Delete(ctx, fmt.Sprintf("session:%s", req.SessionID))
		return nil, domain.ErrSessionExpired
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	_ = s.sessionStore.Delete(ctx, fmt.Sprintf("session:%s", req.SessionID))

	newSessionID := ulid.Make().String()
	newAccessToken, newRefreshToken, hashedToken, err := s.generateTokens(user.Id, session.TenantID, newSessionID)
	if err != nil {
		return nil, err
	}

	newSession := domain.UserSession{
		SessionID:    newSessionID,
		UserID:       user.Id,
		TenantID:     session.TenantID,
		RefreshToken: string(hashedToken),
		ClientIP:     session.ClientIP,
		UserAgent:    session.UserAgent,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	newSessionJSON, err := json.Marshal(newSession)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := s.sessionStore.Set(ctx, fmt.Sprintf("session:%s", newSessionID), string(newSessionJSON), 7*24*time.Hour); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return &UserAuthenticateResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		User: UserResponse{
			Id:                  user.Id,
			Name:                user.Name,
			Username:            user.Username,
			Email:               user.Email,
			ForcePasswordChange: user.ForcePasswordChange,
			OnboardingCompleted: user.OnboardingCompleted,
		},
		Tenant: TenantResponse{
			Id: session.TenantID,
		},
	}, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	if err := s.sessionStore.Delete(ctx, fmt.Sprintf("session:%s", sessionID)); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (s *Service) CompleteOnboarding(ctx context.Context, userID string, tenantID string, req *OnboardingRequest, image *domain.UploadInput) error {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return domain.ErrNotFound
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.ErrNotFound
	}

	var logoS3Key string
	if image != nil {
		key, err := s.fileStorage.PutObject(ctx, *image)
		if err != nil {
			return fmt.Errorf("failed to upload image: %w", err)
		}
		logoS3Key = key
	}

	org := &domain.Organization{
		Id:           ulid.Make().String(),
		TenantId:     tenantID,
		Name:         tenant.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		State:        req.State,
		PostalCode:   req.PostalCode,
		Country:      req.Country,
		TaxId:        req.TaxId,
		LogoS3Key:    logoS3Key,
	}

	if err := s.orgRepo.Insert(ctx, org); err != nil {
		return fmt.Errorf("failed to save organization: %w", err)
	}

	user.OnboardingCompleted = true
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to complete onboarding: %w", err)
	}

	return nil
}

func (s *Service) ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error {
	identifier := base64.StdEncoding.EncodeToString([]byte(req.Email))
	referenceId := fmt.Sprintf("%s-%s", identifier, ulid.Make().String())

	otp, err := randutil.GenerateRandomInteger(otpLength)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	if err := s.sessionStore.Set(ctx, fmt.Sprintf("otp-%s", referenceId), otp, otpExpiry); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		url := fmt.Sprintf("%s/reset-password?ref=%s&token=%s", s.addr, referenceId, otp)
		if err := s.mailer.SendResetPassword(req.Email, mail.ResetPasswordData{Name: req.Email, URL: url}); err != nil {
			log.Printf("failed to send reset password: %v", err)
		}
	}()

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	trimmedRef := strings.TrimSpace(req.ReferenceId)
	if err := s.verifyOTP(ctx, trimmedRef, req.Token); err != nil {
		return err
	}

	parts := strings.SplitN(req.ReferenceId, "-", 2)
	if len(parts) < 2 {
		return domain.ErrBadRequest
	}

	emailBytes, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("failed to decode email: %w", err)
	}

	user, err := s.userRepo.GetByEmail(ctx, string(emailBytes))
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *Service) verifyOTP(ctx context.Context, referenceID string, otp string) error {
	trimmedRef := strings.TrimSpace(referenceID)
	trimmedOTP := strings.TrimSpace(otp)

	storedOTP, err := s.sessionStore.Get(ctx, fmt.Sprintf("otp-%s", trimmedRef))
	if err != nil {
		return domain.ErrOTPInvalid
	}

	if storedOTP != trimmedOTP {
		return domain.ErrOTPInvalid
	}

	if err := s.sessionStore.Delete(ctx, fmt.Sprintf("otp-%s", trimmedRef)); err != nil {
		return fmt.Errorf("failed to delete otp: %w", err)
	}

	return nil
}

func (s *Service) generateTokens(userID, tenantID, sessionID string) (string, string, []byte, error) {
	accessToken, err := s.tokenGenerator.CreateAccessToken(userID, tenantID, sessionID, 15*time.Minute)
	if err != nil {
		return "", "", nil, err
	}

	refreshToken, err := s.tokenGenerator.GenerateRefreshToken()
	if err != nil {
		return "", "", nil, err
	}

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, hashedToken, nil
}

func generateSlug(name string, id string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "workspace"
	}
	suffix := strings.ToLower(id[:8])
	return fmt.Sprintf("%s-%s", slug, suffix)
}
