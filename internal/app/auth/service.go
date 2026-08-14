package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/parxyws/cozybox/internal/config"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/pkg/jwt"
	"github.com/parxyws/cozybox/internal/pkg/mail"
	"github.com/parxyws/cozybox/internal/pkg/randutil"
	"github.com/parxyws/cozybox/internal/store"
	"golang.org/x/crypto/bcrypt"
)

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

const (
	otpLength = 6
	otpExpiry = 2 * time.Minute
)

type AuthService interface {
	Register(ctx context.Context, req RegisterUserRequest) (*RegisterResponse, error)
	VerifyEmail(ctx context.Context, req VerifyEmailRequest) (*VerifyEmailResponse, error)
	Login(ctx context.Context, req LoginRequest) (*UserAuthenticateResponse, error)
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (*UserAuthenticateResponse, error)
	ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (*RegisterResponse, error)
	ResetPassword(ctx context.Context, req ResetPasswordRequest) error
	Logout(ctx context.Context, token string) error
	ListWorkspaces(ctx context.Context, userID string) ([]WorkspaceResponse, error)
	SwitchWorkspace(ctx context.Context, userID string, req SwitchWorkspaceRequest) (*UserAuthenticateResponse, error)
	CompleteOnboarding(ctx context.Context, userID string, tenantID string, req OnboardingRequest, logoFile []byte, logoFilename string) (*WorkspaceResponse, error)
	Shutdown() error
}

type authService struct {
	appStore       store.Store
	sessionStore   SessionStore
	tokenGenerator jwt.TokenGenerator
	mailer         Mailer
	fileStorage    FileStorage
	addr           string
	wg             sync.WaitGroup
}

func NewAuthService(
	appStore store.Store,
	sessionStore SessionStore,
	tokenGenerator jwt.TokenGenerator,
	mailer Mailer,
	fileStorage FileStorage,
	addr string,
) AuthService {
	return &authService{
		appStore:       appStore,
		sessionStore:   sessionStore,
		tokenGenerator: tokenGenerator,
		mailer:         mailer,
		fileStorage:    fileStorage,
		addr:           addr,
	}
}

func (s *authService) Shutdown() error {
	s.wg.Wait()
	return nil
}

func (s *authService) Register(ctx context.Context, req RegisterUserRequest) (*RegisterResponse, error) {
	currentTime := time.Now()
	identifier := base64.StdEncoding.EncodeToString([]byte(req.Email))
	referenceId := fmt.Sprintf("%s-%s", identifier, ulid.Make().String())
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password hash: %w", err)
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

	tenantId := ulid.Make().String()
	slug := generateSlug(req.TenantName, tenantId)
	tenant := &domain.Tenant{
		Id:        tenantId,
		Name:      req.TenantName,
		Slug:      slug,
		Status:    domain.TenantStatusActive,
		Type:      domain.WorkspacePersonal,
		OwnerId:   user.Id,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	member := &domain.TenantMember{
		Id:       ulid.Make().String(),
		TenantId: tenant.Id,
		UserId:   user.Id,
		Role:     domain.TenantRoleOwner,
		JoinedAt: currentTime,
	}

	if err := s.appStore.Tenant().RegisterTenantOwner(ctx, user, tenant, member, nil); err != nil {
		return nil, fmt.Errorf("failed to complete registration: %w", err)
	}

	marshaledData, err := json.Marshal(UserResponse{
		Id:       user.Id,
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to serialize user data: %w", err)
	}

	otp, err := randutil.GenerateRandomInteger(otpLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	if err := s.sessionStore.SetMultiple(ctx, map[string]string{
		fmt.Sprintf("ref-%s", referenceId): string(marshaledData),
		fmt.Sprintf("otp-%s", referenceId): otp,
	}, otpExpiry); err != nil {
		return nil, fmt.Errorf("failed to store registration session: %w", err)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.mailer.SendOTP(req.Email, mail.OTPData{Name: req.Name, OTP: otp}); err != nil {
			fmt.Fprintf(os.Stderr, "failed to send otp to %s: %v\n", req.Email, err)
		}
	}()

	return &RegisterResponse{ReferenceId: referenceId}, nil
}

func (s *authService) VerifyEmail(ctx context.Context, req VerifyEmailRequest) (*VerifyEmailResponse, error) {
	trimmedRef := strings.TrimSpace(req.ReferenceId)

	if err := s.verifyOTP(ctx, trimmedRef, req.Otp); err != nil {
		return nil, domain.ErrInvalidOTP
	}

	data, err := s.sessionStore.Get(ctx, fmt.Sprintf("ref-%s", trimmedRef))
	if err != nil {
		return nil, domain.ErrSessionExpired
	}

	var userResp UserResponse
	if err := json.Unmarshal([]byte(data), &userResp); err != nil {
		return nil, fmt.Errorf("failed to parse user session: %w", err)
	}

	existingUser, err := s.appStore.User().GetUserByID(ctx, userResp.Id)
	if err != nil {
		return nil, err
	}
	existingUser.IsVerified = true
	existingUser.UpdatedAt = time.Now()
	if err := s.appStore.User().UpdateUser(ctx, existingUser); err != nil {
		return nil, err
	}

	return &VerifyEmailResponse{ReferenceId: trimmedRef}, nil
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*UserAuthenticateResponse, error) {
	user, err := s.appStore.User().GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, domain.ErrInvalidCredential
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidCredential
	}

	if !user.IsVerified {
		return nil, domain.ErrEmailNotVerified
	}

	memberships, err := s.appStore.Tenant().ListTenantMembersByUserID(ctx, user.Id)
	if err != nil {
		return nil, err
	}

	if len(memberships) == 0 {
		return nil, domain.ErrNotFound
	}

	member := memberFromList(memberships, "")
	sessionId := ulid.Make().String()
	accessToken, refreshToken, hashedToken, err := s.generateTokens(user.Id, member.TenantId, sessionId)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	session := domain.UserSession{
		SessionID:    sessionId,
		UserID:       user.Id,
		TenantID:     member.TenantId,
		TenantName:   member.Tenant.Name,
		TenantSlug:   member.Tenant.Slug,
		TenantRole:   string(member.Role),
		TenantType:   string(member.Tenant.Type),
		RefreshToken: string(hashedToken),
		ClientIP:     req.ClientIP,
		UserAgent:    req.UserAgent,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize session: %w", err)
	}

	if err := s.sessionStore.Set(ctx, fmt.Sprintf("session:%s", sessionId), string(sessionJSON), 7*24*time.Hour); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	user.LastLogin = time.Now()
	_ = s.appStore.User().UpdateUser(ctx, user)

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
		Tenant: WorkspaceResponse{
			Id:   member.Tenant.Id,
			Name: member.Tenant.Name,
			Slug: member.Tenant.Slug,
			Role: string(member.Role),
			Type: string(member.Tenant.Type),
		},
		ActiveWorkspace: member.Tenant.Id,
		Workspaces:      s.listWorkspacesFromMembers(memberships),
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*UserAuthenticateResponse, error) {
	sessionJSON, err := s.sessionStore.Get(ctx, fmt.Sprintf("session:%s", req.SessionID))
	if err != nil {
		return nil, domain.ErrSessionExpired
	}

	var session domain.UserSession
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return nil, fmt.Errorf("failed to parse session: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.sessionStore.Delete(ctx, fmt.Sprintf("session:%s", req.SessionID))
		return nil, domain.ErrSessionExpired
	}

	if err := bcrypt.CompareHashAndPassword([]byte(session.RefreshToken), []byte(req.RefreshToken)); err != nil {
		return nil, domain.ErrInvalidToken
	}

	user, err := s.appStore.User().GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	memberships, err := s.appStore.Tenant().ListTenantMembersByUserID(ctx, user.Id)
	if err != nil {
		return nil, err
	}

	newAccessToken, newRefreshToken, newHashedToken, err := s.generateTokens(user.Id, session.TenantID, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	session.RefreshToken = string(newHashedToken)
	session.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)

	updatedSessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize session: %w", err)
	}

	if err := s.sessionStore.Set(ctx, fmt.Sprintf("session:%s", req.SessionID), string(updatedSessionJSON), 7*24*time.Hour); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	activeMember := memberFromList(memberships, session.TenantID)

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
		Tenant: WorkspaceResponse{
			Id:   activeMember.Tenant.Id,
			Name: activeMember.Tenant.Name,
			Slug: activeMember.Tenant.Slug,
			Role: string(activeMember.Role),
			Type: string(activeMember.Tenant.Type),
		},
		ActiveWorkspace: activeMember.Tenant.Id,
		Workspaces:      s.listWorkspacesFromMembers(memberships),
	}, nil
}

func (s *authService) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (*RegisterResponse, error) {
	user, err := s.appStore.User().GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	identifier := base64.StdEncoding.EncodeToString([]byte(user.Email))
	referenceId := fmt.Sprintf("%s-%s", identifier, ulid.Make().String())
	token, err := randutil.GenerateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate reset token: %w", err)
	}

	if err := s.sessionStore.SetMultiple(ctx, map[string]string{
		fmt.Sprintf("pw-reset-ref-%s", referenceId):   user.Email,
		fmt.Sprintf("pw-reset-token-%s", referenceId): token,
	}, 15*time.Minute); err != nil {
		return nil, fmt.Errorf("failed to store password reset session: %w", err)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		resetLink := fmt.Sprintf("%s/reset-password?reference_id=%s&token=%s", s.addr, referenceId, token)
		if err := s.mailer.SendResetPassword(user.Email, mail.ResetPasswordData{Name: user.Name, URL: resetLink}); err != nil {
			fmt.Fprintf(os.Stderr, "failed to send reset password email to %s: %v\n", user.Email, err)
		}
	}()

	return &RegisterResponse{ReferenceId: referenceId}, nil
}

func (s *authService) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	storedToken, err := s.sessionStore.Get(ctx, fmt.Sprintf("pw-reset-token-%s", req.ReferenceId))
	if err != nil || storedToken != req.Token {
		return domain.ErrInvalidToken
	}

	storedEmail, err := s.sessionStore.Get(ctx, fmt.Sprintf("pw-reset-ref-%s", req.ReferenceId))
	if err != nil || storedEmail != req.Email {
		return domain.ErrInvalidToken
	}

	user, err := s.appStore.User().GetUserByEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.Password = string(hashedPassword)
	user.ForcePasswordChange = false
	user.UpdatedAt = time.Now()

	if err := s.appStore.User().UpdateUser(ctx, user); err != nil {
		return err
	}

	_ = s.sessionStore.Delete(ctx, fmt.Sprintf("pw-reset-token-%s", req.ReferenceId), fmt.Sprintf("pw-reset-ref-%s", req.ReferenceId))
	return nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	claims, err := s.tokenGenerator.VerifyAccessToken(token)
	if err != nil {
		return domain.ErrInvalidToken
	}
	return s.sessionStore.Delete(ctx, fmt.Sprintf("session:%s", claims.SessionID))
}

func (s *authService) ListWorkspaces(ctx context.Context, userID string) ([]WorkspaceResponse, error) {
	memberships, err := s.appStore.Tenant().ListTenantMembersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.listWorkspacesFromMembers(memberships), nil
}

func (s *authService) SwitchWorkspace(ctx context.Context, userID string, req SwitchWorkspaceRequest) (*UserAuthenticateResponse, error) {
	memberships, err := s.appStore.Tenant().ListTenantMembersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var targetMember *domain.TenantMember
	for _, m := range memberships {
		if m.TenantId == req.WorkspaceID {
			targetMember = &m
			break
		}
	}

	if targetMember == nil {
		return nil, domain.ErrUnauthorized
	}

	user, err := s.appStore.User().GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	sessionId := ulid.Make().String()
	accessToken, refreshToken, hashedToken, err := s.generateTokens(user.Id, targetMember.TenantId, sessionId)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	session := domain.UserSession{
		SessionID:    sessionId,
		UserID:       user.Id,
		TenantID:     targetMember.TenantId,
		TenantName:   targetMember.Tenant.Name,
		TenantSlug:   targetMember.Tenant.Slug,
		TenantRole:   string(targetMember.Role),
		TenantType:   string(targetMember.Tenant.Type),
		RefreshToken: string(hashedToken),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize session: %w", err)
	}

	if err := s.sessionStore.Set(ctx, fmt.Sprintf("session:%s", sessionId), string(sessionJSON), 7*24*time.Hour); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
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
		Tenant: WorkspaceResponse{
			Id:   targetMember.Tenant.Id,
			Name: targetMember.Tenant.Name,
			Slug: targetMember.Tenant.Slug,
			Role: string(targetMember.Role),
			Type: string(targetMember.Tenant.Type),
		},
		ActiveWorkspace: targetMember.Tenant.Id,
		Workspaces:      s.listWorkspacesFromMembers(memberships),
	}, nil
}

func (s *authService) CompleteOnboarding(ctx context.Context, userID string, tenantID string, req OnboardingRequest, logoFile []byte, logoFilename string) (*WorkspaceResponse, error) {
	user, err := s.appStore.User().GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	tenant, err := s.appStore.Tenant().GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	org, err := s.appStore.Organization().GetOrganizationByTenantID(ctx, tenantID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}

	var logoS3Key string
	if len(logoFile) > 0 {
		key, err := s.fileStorage.PutObject(ctx, domain.UploadInput{
			Object:      bytes.NewReader(logoFile),
			ObjectName:  fmt.Sprintf("tenants/%s/logos/%s", tenantID, logoFilename),
			ObjectSize:  int64(len(logoFile)),
			ContentType: "image/png",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to upload organization logo: %w", err)
		}
		logoS3Key = key
	}

	now := time.Now()
	if org == nil {
		org = &domain.Organization{
			Id:              ulid.Make().String(),
			TenantId:        tenantID,
			Name:            req.TenantName,
			Email:           req.Email,
			Phone:           req.Phone,
			AddressLine1:    req.AddressLine1,
			AddressLine2:    req.AddressLine2,
			City:            req.City,
			State:           req.State,
			PostalCode:      req.PostalCode,
			Country:         req.Country,
			TaxId:           req.TaxId,
			LogoS3Key:       logoS3Key,
			Website:         req.Website,
			Timezone:        req.Timezone,
			DefaultCurrency: req.DefaultCurrency,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := s.appStore.Organization().InsertOrganization(ctx, org); err != nil {
			return nil, fmt.Errorf("failed to create organization profile: %w", err)
		}
	} else {
		org.Name = req.TenantName
		if req.Email != "" {
			org.Email = req.Email
		}
		org.Phone = req.Phone
		org.AddressLine1 = req.AddressLine1
		org.AddressLine2 = req.AddressLine2
		org.City = req.City
		org.State = req.State
		org.PostalCode = req.PostalCode
		org.Country = req.Country
		org.TaxId = req.TaxId
		if logoS3Key != "" {
			org.LogoS3Key = logoS3Key
		}
		org.Website = req.Website
		org.Timezone = req.Timezone
		org.DefaultCurrency = req.DefaultCurrency
		org.UpdatedAt = now
		if err := s.appStore.Organization().UpdateOrganization(ctx, org); err != nil {
			return nil, fmt.Errorf("failed to update organization profile: %w", err)
		}
	}

	user.OnboardingCompleted = true
	user.UpdatedAt = now
	if err := s.appStore.User().UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user onboarding status: %w", err)
	}

	tenant.Name = req.TenantName
	tenant.UpdatedAt = now
	if err := s.appStore.Tenant().UpdateTenant(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant name: %w", err)
	}

	return &WorkspaceResponse{
		Id:   tenant.Id,
		Name: tenant.Name,
		Slug: tenant.Slug,
		Type: string(tenant.Type),
	}, nil
}

func (s *authService) verifyOTP(ctx context.Context, referenceId, otp string) error {
	storedOTP, err := s.sessionStore.Get(ctx, fmt.Sprintf("otp-%s", referenceId))
	if err != nil {
		return domain.ErrInvalidOTP
	}
	if storedOTP != otp {
		return domain.ErrInvalidOTP
	}
	_ = s.sessionStore.Delete(ctx, fmt.Sprintf("otp-%s", referenceId))
	return nil
}

func (s *authService) generateTokens(userID, tenantID, sessionID string) (accessToken, refreshToken string, hashedToken []byte, err error) {
	accessToken, err = s.tokenGenerator.CreateAccessToken(userID, tenantID, sessionID, config.AccessTokenDuration)
	if err != nil {
		return "", "", nil, err
	}
	refreshToken, err = s.tokenGenerator.GenerateRefreshToken()
	if err != nil {
		return "", "", nil, err
	}
	hashedToken, err = bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return "", "", nil, err
	}
	return accessToken, refreshToken, hashedToken, nil
}

func (s *authService) listWorkspacesFromMembers(memberships []domain.TenantMember) []WorkspaceResponse {
	res := make([]WorkspaceResponse, 0, len(memberships))
	for _, m := range memberships {
		res = append(res, WorkspaceResponse{
			Id:   m.Tenant.Id,
			Name: m.Tenant.Name,
			Slug: m.Tenant.Slug,
			Role: string(m.Role),
			Type: string(m.Tenant.Type),
		})
	}
	return res
}

func generateSlug(name, tenantId string) string {
	slug := strings.ToLower(name)
	slug = slugRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if len(tenantId) >= 8 {
		slug = fmt.Sprintf("%s-%s", slug, tenantId[:8])
	}
	return slug
}

func memberFromList(members []domain.TenantMember, preferredTenantID string) domain.TenantMember {
	if preferredTenantID != "" {
		for _, m := range members {
			if m.TenantId == preferredTenantID {
				return m
			}
		}
	}
	return members[0]
}
