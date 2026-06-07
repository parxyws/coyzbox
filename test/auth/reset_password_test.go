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

func TestResetPassword_Success(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	refID := "dGVzdEBleGFtcGxlLmNvbQ==-01JMOCK-Z"

	m.sessionStore.On("Get", ctx, "otp-"+refID).Return("123456", nil)
	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)
	m.userRepo.On("GetByEmail", ctx, "test@example.com").Return(&domain.User{Id: "user-123", Email: "test@example.com"}, nil)
	m.userRepo.On("Update", ctx, mock.Anything).Return(nil)

	req := &auth.ResetPasswordRequest{
		Email:       "test@example.com",
		ReferenceId: refID,
		Token:       "123456",
		Password:    "newpass123",
	}

	err := svc.ResetPassword(ctx, req)
	require.NoError(t, err)
}

func TestResetPassword_InvalidOTP(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	refID := "dGVzdEBleGFtcGxlLmNvbQ==-01JMOCK-Z"

	m.sessionStore.On("Get", ctx, "otp-"+refID).Return("999999", nil)
	// Note: Delete is NOT called when OTP mismatches — the service returns before reaching Delete

	req := &auth.ResetPasswordRequest{
		Email:       "test@example.com",
		ReferenceId: refID,
		Token:       "123456",
		Password:    "newpass123",
	}

	err := svc.ResetPassword(ctx, req)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrOTPInvalid)
}

func TestResetPassword_BadReferenceFormat(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	refID := "nodasheshere"

	m.sessionStore.On("Get", ctx, "otp-"+refID).Return("123456", nil)
	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)

	req := &auth.ResetPasswordRequest{
		Email:       "test@example.com",
		ReferenceId: refID,
		Token:       "123456",
		Password:    "newpass123",
	}

	err := svc.ResetPassword(ctx, req)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrBadRequest)
}

func TestResetPassword_Base64DecodeFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	refID := "!!!invalid-base64!!!--01JMOCK-Z"

	m.sessionStore.On("Get", ctx, "otp-"+refID).Return("123456", nil)
	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)

	req := &auth.ResetPasswordRequest{
		Email:       "test@example.com",
		ReferenceId: refID,
		Token:       "123456",
		Password:    "newpass123",
	}

	err := svc.ResetPassword(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode email")
}

func TestResetPassword_UserNotFound(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	refID := "dGVzdEBleGFtcGxlLmNvbQ==-01JMOCK-Z"

	m.sessionStore.On("Get", ctx, "otp-"+refID).Return("123456", nil)
	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)
	m.userRepo.On("GetByEmail", ctx, "test@example.com").Return(nil, domain.ErrNotFound)

	req := &auth.ResetPasswordRequest{
		Email:       "test@example.com",
		ReferenceId: refID,
		Token:       "123456",
		Password:    "newpass123",
	}

	err := svc.ResetPassword(ctx, req)
	assert.Error(t, err)
}

func TestResetPassword_UpdateFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	refID := "dGVzdEBleGFtcGxlLmNvbQ==-01JMOCK-Z"

	m.sessionStore.On("Get", ctx, "otp-"+refID).Return("123456", nil)
	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)
	m.userRepo.On("GetByEmail", ctx, "test@example.com").Return(&domain.User{Id: "user-123", Email: "test@example.com"}, nil)
	m.userRepo.On("Update", ctx, mock.Anything).Return(errors.New("db error"))

	req := &auth.ResetPasswordRequest{
		Email:       "test@example.com",
		ReferenceId: refID,
		Token:       "123456",
		Password:    "newpass123",
	}

	err := svc.ResetPassword(ctx, req)
	assert.Error(t, err)
}
