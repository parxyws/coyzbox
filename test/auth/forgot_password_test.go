package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/parxyws/cozybox/internal/app/auth"
)

func TestForgotPassword_Success(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.ForgotPasswordRequest{Email: "test@example.com"}

	m.sessionStore.On("Set", ctx, mock.MatchedBy(func(key string) bool {
		return len(key) >= 4 && key[:4] == "otp-"
	}), mock.Anything, mock.Anything).Return(nil)

	m.mailer.On("SendResetPassword", req.Email, mock.Anything).Return(nil).Maybe()

	err := svc.ForgotPassword(ctx, req)
	require.NoError(t, err)
	m.sessionStore.AssertExpectations(t)
}

func TestForgotPassword_SessionStoreFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	req := &auth.ForgotPasswordRequest{Email: "test@example.com"}

	m.sessionStore.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("redis error"))

	err := svc.ForgotPassword(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store OTP")
}
