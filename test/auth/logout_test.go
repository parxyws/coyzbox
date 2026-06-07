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

func TestLogout_Success(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(nil)

	err := svc.Logout(ctx, "session-123")
	require.NoError(t, err)
}

func TestLogout_DeleteFails(t *testing.T) {
	svc, m := newServiceWithMocks(t)
	ctx := context.Background()

	m.sessionStore.On("Delete", mock.Anything, mock.Anything).Return(errors.New("redis error"))

	err := svc.Logout(ctx, "session-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete session")
}

var _ = auth.OnboardingRequest{}
