package config

import "time"

type ContextKey string

const (
	RequestID ContextKey = "request_id"
	UserID    ContextKey = "user_id"
	TenantID  ContextKey = "tenant_id"
	SessionID ContextKey = "session_id"
)

const (
	CtxTimeout           = 3
	AccessTokenDuration  = 24 * time.Hour
	RefreshTokenDuration = 7 * 24 * time.Hour
)
