package auth_test

import (
	"context"
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

func hashPassword(pwd string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	return string(h)
}

var correctHash = hashPassword("correctpass")

func TestLogin_Success(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.LoginRequest{Email: "test@example.com", Password: "correctpass"}

	m.userRepo.On("GetByEmail", ctx, req.Email).Return(&domain.User{
		Id:         "user-123",
		Email:      req.Email,
		Password:   correctHash,
		IsVerified: true,
	}, nil)

	m.memberRepo.On("ListByUserID", ctx, "user-123").Return([]domain.TenantMember{{
		Id:       "member-1",
		TenantId: "tenant-1",
		UserId:   "user-123",
		Role:     domain.TenantRoleOwner,
		Tenant:   domain.Tenant{Id: "tenant-1", Name: "Test Corp", Slug: "test-corp"},
	}}, nil)

	m.sessionStore.On("Set", ctx, mock.Anything, mock.Anything, 7*24*time.Hour).Return(nil)
	m.userRepo.On("Update", ctx, mock.Anything).Return(nil)

	resp, err := svc.Login(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "access-token", resp.AccessToken)
	assert.Equal(t, "refresh-token", resp.RefreshToken)
	assert.Equal(t, "test@example.com", resp.User.Email)
	assert.Equal(t, "tenant-1", resp.Tenant.Id)
}

func TestLogin_UserNotFound(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	m.userRepo.On("GetByEmail", ctx, "notfound@example.com").Return(nil, domain.ErrNotFound)

	resp, err := svc.Login(ctx, &auth.LoginRequest{Email: "notfound@example.com", Password: "anything"})
	assert.ErrorIs(t, err, domain.ErrInvalidCredential)
	assert.Nil(t, resp)
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.LoginRequest{Email: "test@example.com", Password: "wrongpass"}

	m.userRepo.On("GetByEmail", ctx, req.Email).Return(&domain.User{
		Id: "user-123", Email: req.Email, Password: correctHash, IsVerified: true,
	}, nil)

	resp, err := svc.Login(ctx, req)
	assert.ErrorIs(t, err, domain.ErrInvalidCredential)
	assert.Nil(t, resp)
}

func TestLogin_EmailNotVerified(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.LoginRequest{Email: "test@example.com", Password: "correctpass"}

	m.userRepo.On("GetByEmail", ctx, req.Email).Return(&domain.User{
		Id: "user-123", Email: req.Email, Password: correctHash, IsVerified: false,
	}, nil)

	resp, err := svc.Login(ctx, req)
	assert.ErrorIs(t, err, domain.ErrEmailNotVerified)
	assert.Nil(t, resp)
}

func TestLogin_MemberNotFound(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.LoginRequest{Email: "test@example.com", Password: "correctpass"}

	m.userRepo.On("GetByEmail", ctx, req.Email).Return(&domain.User{
		Id: "user-123", Email: req.Email, Password: correctHash, IsVerified: true,
	}, nil)
	m.memberRepo.On("ListByUserID", ctx, "user-123").Return([]domain.TenantMember(nil), domain.ErrNotFound)

	resp, err := svc.Login(ctx, req)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, resp)
}

func TestLogin_TokenGenerationFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()
	m.tokenGenerator.Err = errors.New("jwt error")

	req := &auth.LoginRequest{Email: "test@example.com", Password: "correctpass"}

	m.userRepo.On("GetByEmail", ctx, req.Email).Return(&domain.User{
		Id: "user-123", Email: req.Email, Password: correctHash, IsVerified: true,
	}, nil)
	m.memberRepo.On("ListByUserID", ctx, "user-123").Return([]domain.TenantMember{{
		Id: "member-1", TenantId: "tenant-1", UserId: "user-123", Role: domain.TenantRoleOwner,
		Tenant: domain.Tenant{Id: "tenant-1", Name: "Test Corp", Slug: "test-corp"},
	}}, nil)

	resp, err := svc.Login(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestLogin_SessionStoreFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.LoginRequest{Email: "test@example.com", Password: "correctpass"}

	m.userRepo.On("GetByEmail", ctx, req.Email).Return(&domain.User{
		Id: "user-123", Email: req.Email, Password: correctHash, IsVerified: true,
	}, nil)
	m.memberRepo.On("ListByUserID", ctx, "user-123").Return([]domain.TenantMember{{
		Id: "member-1", TenantId: "tenant-1", UserId: "user-123", Role: domain.TenantRoleOwner,
		Tenant: domain.Tenant{Id: "tenant-1", Name: "Test Corp", Slug: "test-corp"},
	}}, nil)
	m.sessionStore.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("redis error"))

	resp, err := svc.Login(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store session")
	assert.Nil(t, resp)
}

func TestLogin_LastLoginUpdateFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.LoginRequest{Email: "test@example.com", Password: "correctpass"}

	m.userRepo.On("GetByEmail", ctx, req.Email).Return(&domain.User{
		Id: "user-123", Email: req.Email, Password: correctHash, IsVerified: true,
	}, nil)
	m.memberRepo.On("ListByUserID", ctx, "user-123").Return([]domain.TenantMember{{
		Id: "member-1", TenantId: "tenant-1", UserId: "user-123", Role: domain.TenantRoleOwner,
		Tenant: domain.Tenant{Id: "tenant-1", Name: "Test Corp", Slug: "test-corp"},
	}}, nil)
	m.sessionStore.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.userRepo.On("Update", ctx, mock.Anything).Return(errors.New("db error"))

	resp, err := svc.Login(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}
