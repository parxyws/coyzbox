package domain

import "time"

type Claims struct {
	UserID    string
	TenantID  string
	SessionID string
}

type TokenGenerator interface {
	CreateAccessToken(userID string, tenantID string, sessionID string, duration time.Duration) (string, error)
	VerifyAccessToken(token string) (*Claims, error)
	GenerateRefreshToken() (string, error)
}
