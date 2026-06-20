package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/parxyws/cozybox/internal/domain"
)

// TokenGenerator defines the interface for creating and verifying tokens.
type TokenGenerator interface {
	CreateAccessToken(userID string, tenantID string, sessionID string, duration time.Duration) (string, error)
	VerifyAccessToken(token string) (*domain.Claims, error)
	GenerateRefreshToken() (string, error)
}

const (
	issuer   = "cozybox"
	audience = "cozybox-api"
)

type customClaims struct {
	UserID    string `json:"user_id"`
	TenantID  string `json:"tenant_id"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

type JWTMaker struct {
	secretKey []byte
}

func NewJWTMaker(secretKey string) (TokenGenerator, error) {
	if len(secretKey) < 32 {
		return nil, errors.New("invalid key size: must be at least 32 characters")
	}
	return &JWTMaker{secretKey: []byte(secretKey)}, nil
}

func (m *JWTMaker) CreateAccessToken(userID, tenantID, sessionID string, duration time.Duration) (string, error) {
	claims := customClaims{
		UserID:    userID,
		TenantID:  tenantID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   userID,
			Audience:  []string{audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        sessionID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *JWTMaker) VerifyAccessToken(tokenStr string) (*domain.Claims, error) {
	keyFunc := func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("invalid signing method")
		}
		return m.secretKey, nil
	}

	jwtToken, err := jwt.ParseWithClaims(tokenStr, &customClaims{}, keyFunc,
		jwt.WithLeeway(5*time.Second),
		jwt.WithIssuer(issuer),
		jwt.WithAudience(audience),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}

	if !jwtToken.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := jwtToken.Claims.(*customClaims)
	if !ok {
		return nil, errors.New("invalid claims type")
	}

	return &domain.Claims{
		UserID:    claims.UserID,
		TenantID:  claims.TenantID,
		SessionID: claims.SessionID,
	}, nil
}

func (m *JWTMaker) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
