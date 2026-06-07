package domain

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource already exists")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrBadRequest        = errors.New("bad request")
	ErrInternalServer    = errors.New("internal server error")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrEmailNotVerified  = errors.New("email not verified")
	ErrOTPInvalid        = errors.New("invalid or expired OTP")
	ErrSessionExpired    = errors.New("session expired")
)
