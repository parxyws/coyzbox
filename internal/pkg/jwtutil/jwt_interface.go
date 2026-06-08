package jwtutil

import (
	"time"

	"github.com/parxyws/cozybox/internal/domain"
)

type TokenGenerator interface {
	CreateAccessToken(userID string, tenantID string, sessionID string, duration time.Duration) (string, error)
	VerifyAccessToken(token string) (*domain.Claims, error)
	GenerateRefreshToken() (string, error)
}
