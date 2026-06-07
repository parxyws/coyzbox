package config

type ContextKey string

const (
	RequestID ContextKey = "request_id"
	UserID    ContextKey = "user_id"
	TenantID  ContextKey = "tenant_id"
	SessionID ContextKey = "session_id"
)
