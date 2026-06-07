package domain

import "time"

type UserSession struct {
	SessionID    string    `json:"session_id"`
	UserID       string    `json:"user_id"`
	TenantID     string    `json:"tenant_id"`
	RefreshToken string    `json:"refresh_token"`
	ClientIP     string    `json:"client_ip"`
	UserAgent    string    `json:"user_agent"`
	ExpiresAt    time.Time `json:"expires_at"`
}
