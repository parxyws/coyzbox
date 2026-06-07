package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/parxyws/cozybox/internal/app/auth"
	"github.com/parxyws/cozybox/internal/domain"
)

func TestRegister_Success(t *testing.T) {
	svc, m := newServiceWithMocks(t)

	req := &auth.RegisterUserRequest{
		Name:       "Test User",
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "securepass123",
		TenantName: "Test Corp",
	}

	m.userRepo.On("Insert", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == req.Email && u.Name == req.Name
	})).Return(nil)

	m.tenantRepo.On("Insert", mock.Anything, mock.MatchedBy(func(tn *domain.Tenant) bool {
		return tn.Name == req.TenantName
	})).Return(nil)

	m.memberRepo.On("Insert", mock.Anything, mock.MatchedBy(func(tm *domain.TenantMember) bool {
		return tm.Role == domain.TenantRoleOwner
	})).Return(nil)

	m.sessionStore.On("SetMultiple", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	m.mailer.On("SendOTP", req.Email, mock.Anything).Return(nil).Maybe()

	resp, err := svc.Register(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.ReferenceId)
	m.userRepo.AssertExpectations(t)
	m.tenantRepo.AssertExpectations(t)
	m.memberRepo.AssertExpectations(t)
	m.sessionStore.AssertExpectations(t)
}

func TestRegister_UserInsertFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)

	req := &auth.RegisterUserRequest{
		Name:       "Test User",
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "securepass123",
		TenantName: "Test Corp",
	}

	m.userRepo.On("Insert", mock.Anything, mock.Anything).Return(errors.New("db error"))

	resp, err := svc.Register(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to register user")
	m.userRepo.AssertExpectations(t)
}

func TestRegister_TenantInsertFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)

	req := &auth.RegisterUserRequest{
		Name:       "Test User",
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "securepass123",
		TenantName: "Test Corp",
	}

	m.userRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)
	m.tenantRepo.On("Insert", mock.Anything, mock.Anything).Return(errors.New("tenant db error"))

	resp, err := svc.Register(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to create tenant")
	m.userRepo.AssertExpectations(t)
	m.tenantRepo.AssertExpectations(t)
}

func TestRegister_MemberInsertFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)

	req := &auth.RegisterUserRequest{
		Name:       "Test User",
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "securepass123",
		TenantName: "Test Corp",
	}

	m.userRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)
	m.tenantRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)
	m.memberRepo.On("Insert", mock.Anything, mock.Anything).Return(errors.New("member db error"))

	resp, err := svc.Register(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to create tenant member")
	m.userRepo.AssertExpectations(t)
	m.tenantRepo.AssertExpectations(t)
	m.memberRepo.AssertExpectations(t)
}

func TestRegister_SessionStoreFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)

	req := &auth.RegisterUserRequest{
		Name:       "Test User",
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "securepass123",
		TenantName: "Test Corp",
	}

	m.userRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)
	m.tenantRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)
	m.memberRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)
	m.sessionStore.On("SetMultiple", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("redis error"))

	resp, err := svc.Register(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to store registration data")
	m.userRepo.AssertExpectations(t)
	m.tenantRepo.AssertExpectations(t)
	m.memberRepo.AssertExpectations(t)
	m.sessionStore.AssertExpectations(t)
}
