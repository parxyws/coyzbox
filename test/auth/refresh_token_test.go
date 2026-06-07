package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/parxyws/cozybox/internal/app/auth"
	"github.com/parxyws/cozybox/internal/domain"
)

func TestRefreshToken_Success(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	// Pre-hash a refresh token
	plainRefreshToken := "old-refresh-token"
	hashedToken, _ := bcrypt.GenerateFromPassword([]byte(plainRefreshToken), bcrypt.MinCost)

	session := domain.UserSession{
		SessionID:    "session-old-123",
		UserID:       "user-123",
		TenantID:     "tenant-1",
		RefreshToken: string(hashedToken),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		ClientIP:     "10.0.0.1",
		UserAgent:    "test-agent",
	}
	sessionJSON, _ := json.Marshal(session)

	m.sessionStore.On("Get", ctx, "session:session-old-123").Return(string(sessionJSON), nil)
	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)

	m.userRepo.On("GetByID", ctx, "user-123").Return(&domain.User{
		Id:    "user-123",
		Email: "test@example.com",
	}, nil)

	m.sessionStore.On("Set", ctx, mock.Anything, mock.Anything, 7*24*time.Hour).Return(nil)

	req := &auth.RefreshTokenRequest{
		SessionID:    "session-old-123",
		RefreshToken: plainRefreshToken,
	}

	resp, err := svc.RefreshToken(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "access-token", resp.AccessToken)
	assert.Equal(t, "refresh-token", resp.RefreshToken)
	assert.Equal(t, "test@example.com", resp.User.Email)
}

func TestRefreshToken_SessionNotFound(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	m.sessionStore.On("Get", ctx, "session:session-xxx").Return("", errors.New("not found"))

	req := &auth.RefreshTokenRequest{SessionID: "session-xxx", RefreshToken: "anything"}
	resp, err := svc.RefreshToken(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrSessionExpired)
}

func TestRefreshToken_RefreshTokenMismatch(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	hashedToken, _ := bcrypt.GenerateFromPassword([]byte("correct-refresh-token"), bcrypt.MinCost)
	session := domain.UserSession{
		SessionID:    "session-123",
		UserID:       "user-123",
		TenantID:     "tenant-1",
		RefreshToken: string(hashedToken),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
	sessionJSON, _ := json.Marshal(session)

	m.sessionStore.On("Get", ctx, "session:session-123").Return(string(sessionJSON), nil)

	req := &auth.RefreshTokenRequest{SessionID: "session-123", RefreshToken: "wrong-refresh-token"}
	resp, err := svc.RefreshToken(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestRefreshToken_SessionExpired(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	hashedToken, _ := bcrypt.GenerateFromPassword([]byte("refresh-token"), bcrypt.MinCost)
	session := domain.UserSession{
		SessionID:    "session-expired-123",
		UserID:       "user-123",
		TenantID:     "tenant-1",
		RefreshToken: string(hashedToken),
		ExpiresAt:    time.Now().Add(-1 * time.Hour),
	}
	sessionJSON, _ := json.Marshal(session)

	m.sessionStore.On("Get", ctx, "session:session-expired-123").Return(string(sessionJSON), nil)
	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)

	req := &auth.RefreshTokenRequest{SessionID: "session-expired-123", RefreshToken: "refresh-token"}
	resp, err := svc.RefreshToken(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrSessionExpired)
}

func TestRefreshToken_UserNotFoundAfterSessionValidation(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	hashedToken, _ := bcrypt.GenerateFromPassword([]byte("refresh-token"), bcrypt.MinCost)
	session := domain.UserSession{
		SessionID:    "session-123",
		UserID:       "user-nonexistent",
		TenantID:     "tenant-1",
		RefreshToken: string(hashedToken),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
	sessionJSON, _ := json.Marshal(session)

	m.sessionStore.On("Get", ctx, "session:session-123").Return(string(sessionJSON), nil)
	// Delete is called after GetByID, but GetByID fails first, so Delete never reached
	m.userRepo.On("GetByID", ctx, "user-nonexistent").Return(nil, domain.ErrNotFound)

	req := &auth.RefreshTokenRequest{SessionID: "session-123", RefreshToken: "refresh-token"}
	resp, err := svc.RefreshToken(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestRefreshToken_TokenGenerationFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()
	m.tokenGenerator.Err = errors.New("jwt error")

	hashedToken, _ := bcrypt.GenerateFromPassword([]byte("refresh-token"), bcrypt.MinCost)
	session := domain.UserSession{
		SessionID:    "session-123",
		UserID:       "user-123",
		TenantID:     "tenant-1",
		RefreshToken: string(hashedToken),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
	sessionJSON, _ := json.Marshal(session)

	m.sessionStore.On("Get", ctx, "session:session-123").Return(string(sessionJSON), nil)
	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)
	m.userRepo.On("GetByID", ctx, "user-123").Return(&domain.User{Id: "user-123", Email: "test@example.com"}, nil)

	req := &auth.RefreshTokenRequest{SessionID: "session-123", RefreshToken: "refresh-token"}
	resp, err := svc.RefreshToken(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestRefreshToken_NewSessionStoreFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	hashedToken, _ := bcrypt.GenerateFromPassword([]byte("refresh-token"), bcrypt.MinCost)
	session := domain.UserSession{
		SessionID:    "session-123",
		UserID:       "user-123",
		TenantID:     "tenant-1",
		RefreshToken: string(hashedToken),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
	sessionJSON, _ := json.Marshal(session)

	m.sessionStore.On("Get", ctx, "session:session-123").Return(string(sessionJSON), nil)
	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)
	m.userRepo.On("GetByID", ctx, "user-123").Return(&domain.User{Id: "user-123", Email: "test@example.com"}, nil)
	m.sessionStore.On("Set", ctx, mock.Anything, mock.Anything, 7*24*time.Hour).Return(errors.New("redis error"))

	req := &auth.RefreshTokenRequest{SessionID: "session-123", RefreshToken: "refresh-token"}
	resp, err := svc.RefreshToken(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to store session")
}

func TestRefreshToken_CorruptSessionJSON(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	m.sessionStore.On("Get", ctx, "session:session-123").Return("not-valid-json{{{", nil)

	req := &auth.RefreshTokenRequest{SessionID: "session-123", RefreshToken: "anything"}
	resp, err := svc.RefreshToken(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}
