package auth

import (
	"context"
	"time"

	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/pkg/mail"
)

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
