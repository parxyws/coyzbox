package models

import "time"

// UserSession is stored in Redis for authentication and revocation
type UserSession struct {
	SessionID    string    `json:"session_id"` // Also used as the Redis key
	UserID       string    `json:"user_id"`
	TenantID     string    `json:"tenant_id"`
	RefreshToken string    `json:"refresh_token"` // Should be hashed when stored
	ClientIP     string    `json:"client_ip"`
	UserAgent    string    `json:"user_agent"`
	ExpiresAt    time.Time `json:"expires_at"`
}
