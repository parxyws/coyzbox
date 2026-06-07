package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/parxyws/cozybox/internal/app/auth"
	"github.com/parxyws/cozybox/internal/domain"
)

func TestVerifyEmail_Success(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	refID := "dGVzdEBleGFtcGxlLmNvbQ==-01JMOCK-Z"

	userData := auth.UserResponse{
		Id:    "user-123",
		Name:  "Test User",
		Email: "test@example.com",
	}
	data, _ := json.Marshal(userData)

	m.sessionStore.On("Get", ctx, "otp-"+refID).Return("123456", nil)
	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)
	m.sessionStore.On("Get", ctx, "ref-"+refID).Return(string(data), nil)
	m.userRepo.On("GetByID", ctx, "user-123").Return(&domain.User{
		Id:         "user-123",
		Email:      "test@example.com",
		IsVerified: false,
	}, nil)
	m.userRepo.On("Update", ctx, mock.MatchedBy(func(u *domain.User) bool {
		return u.IsVerified == true
	})).Return(nil)

	req := &auth.VerifyEmailRequest{ReferenceId: refID, Otp: "123456"}
	resp, err := svc.VerifyEmail(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, refID, resp.ReferenceId)
}

func TestVerifyEmail_OTPNotFound(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()
	refID := "dGVzdEBleGFtcGxlLmNvbQ==-01JMOCK-Z"

	m.sessionStore.On("Get", ctx, "otp-"+refID).Return("", errors.New("not found"))

	req := &auth.VerifyEmailRequest{ReferenceId: refID, Otp: "123456"}
	resp, err := svc.VerifyEmail(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrOTPInvalid)
}

func TestVerifyEmail_OTPMismatch(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()
	refID := "dGVzdEBleGFtcGxlLmNvbQ==-01JMOCK-Z"

	m.sessionStore.On("Get", ctx, "otp-"+refID).Return("999999", nil)
	// Delete NOT called when OTP mismatches

	req := &auth.VerifyEmailRequest{ReferenceId: refID, Otp: "123456"}
	resp, err := svc.VerifyEmail(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrOTPInvalid)
}

func TestVerifyEmail_RefDataNotFound(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()
	refID := "dGVzdEBleGFtcGxlLmNvbQ==-01JMOCK-Z"

	m.sessionStore.On("Get", ctx, "otp-"+refID).Return("123456", nil)
	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)
	m.sessionStore.On("Get", ctx, "ref-"+refID).Return("", errors.New("not found"))

	req := &auth.VerifyEmailRequest{ReferenceId: refID, Otp: "123456"}
	resp, err := svc.VerifyEmail(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestVerifyEmail_UserNotFound(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()
	refID := "dGVzdEBleGFtcGxlLmNvbQ==-01JMOCK-Z"

	userData := auth.UserResponse{Id: "user-123", Name: "Test", Email: "test@example.com"}
	data, _ := json.Marshal(userData)

	m.sessionStore.On("Get", ctx, "otp-"+refID).Return("123456", nil)
	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)
	m.sessionStore.On("Get", ctx, "ref-"+refID).Return(string(data), nil)
	m.userRepo.On("GetByID", ctx, "user-123").Return(nil, domain.ErrNotFound)

	req := &auth.VerifyEmailRequest{ReferenceId: refID, Otp: "123456"}
	resp, err := svc.VerifyEmail(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestVerifyEmail_UserUpdateFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()
	refID := "dGVzdEBleGFtcGxlLmNvbQ==-01JMOCK-Z"

	userData := auth.UserResponse{Id: "user-123", Name: "Test", Email: "test@example.com"}
	data, _ := json.Marshal(userData)

	m.sessionStore.On("Get", ctx, "otp-"+refID).Return("123456", nil)
	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)
	m.sessionStore.On("Get", ctx, "ref-"+refID).Return(string(data), nil)
	m.userRepo.On("GetByID", ctx, "user-123").Return(&domain.User{
		Id:         "user-123",
		IsVerified: false,
	}, nil)
	m.userRepo.On("Update", ctx, mock.Anything).Return(errors.New("update error"))

	req := &auth.VerifyEmailRequest{ReferenceId: refID, Otp: "123456"}
	resp, err := svc.VerifyEmail(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestVerifyEmail_WhitespaceTrimRefID(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()
	refID := "dGVzdEBleGFtcGxlLmNvbQ==-01JMOCK-Z"
	trimmedRefID := "dGVzdEBleGFtcGxlLmNvbQ==-01JMOCK-Z"

	userData := auth.UserResponse{Id: "user-123", Name: "Test", Email: "test@example.com"}
	data, _ := json.Marshal(userData)

	m.sessionStore.On("Get", ctx, "otp-"+trimmedRefID).Return("123456", nil)
	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)
	m.sessionStore.On("Get", ctx, "ref-"+trimmedRefID).Return(string(data), nil)
	m.userRepo.On("GetByID", ctx, "user-123").Return(&domain.User{Id: "user-123", IsVerified: false}, nil)
	m.userRepo.On("Update", ctx, mock.Anything).Return(nil)

	req := &auth.VerifyEmailRequest{ReferenceId: " " + refID + " ", Otp: "123456"}
	resp, err := svc.VerifyEmail(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, trimmedRefID, resp.ReferenceId)
}
