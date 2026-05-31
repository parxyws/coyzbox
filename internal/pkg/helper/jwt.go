package helper

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	ISS = "cozybox"
)

type TokenGenerator interface {
	CreateAccessToken(userID string, tenantID string, sessionID string, duration time.Duration) (string, error)
	VerifyAccessToken(token string) (*CustomClaims, error)
	GenerateRefreshToken() (string, error)
}

type JWTMaker struct {
	secretKey []byte
}

type CustomClaims struct {
	UserID    string `json:"user_id"`
	TenantID  string `json:"tenant_id"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

func NewJWTMaker(secretKey string) (TokenGenerator, error) {
	if len(secretKey) < 32 {
		return nil, errors.New("invalid key size: must be at least 32 characters")
	}
	return &JWTMaker{
		secretKey: []byte(secretKey),
	}, nil
}

func (maker *JWTMaker) CreateAccessToken(userID string, tenantID string, sessionID string, duration time.Duration) (string, error) {
	claims := CustomClaims{
		UserID:    userID,
		TenantID:  tenantID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    ISS,
			Subject:   userID,
			Audience:  []string{"cozybox-api"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        sessionID,
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString(maker.secretKey)
}

func (maker *JWTMaker) VerifyAccessToken(token string) (*CustomClaims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("invalid authentication token marker")
		}
		return maker.secretKey, nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &CustomClaims{}, keyFunc,
		jwt.WithLeeway(5*time.Second),
		jwt.WithIssuer(ISS),
		jwt.WithAudience("cozybox-api"),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}

	claims, ok := jwtToken.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid authentication token marker")
	}

	return claims, nil
}

func (maker *JWTMaker) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
