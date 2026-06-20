package domain

import "time"

type UserSession struct {
	SessionID    string    `json:"session_id"`
	UserID       string    `json:"user_id"`
	TenantID     string    `json:"tenant_id"`
	TenantName   string    `json:"tenant_name"`
	TenantSlug   string    `json:"tenant_slug"`
	TenantRole   string    `json:"tenant_role"`
	TenantType   string    `json:"tenant_type"`
	RefreshToken string    `json:"refresh_token"`
	ClientIP     string    `json:"client_ip"`
	UserAgent    string    `json:"user_agent"`
	ExpiresAt    time.Time `json:"expires_at"`
}
