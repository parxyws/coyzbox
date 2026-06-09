package jwtutil

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {
	secretKey := "super_secret_key_minimum_32_characters_long"
	shortKey := "short_key"

	t.Run("NewJWTMaker - validation", func(t *testing.T) {
		maker, err := NewJWTMaker(secretKey)
		require.NoError(t, err)
		require.NotNil(t, maker)

		makerShort, err := NewJWTMaker(shortKey)
		require.Error(t, err)
		require.Nil(t, makerShort)
		require.Contains(t, err.Error(), "must be at least 32 characters")
	})

	t.Run("Create and Verify Token - success", func(t *testing.T) {
		maker, err := NewJWTMaker(secretKey)
		require.NoError(t, err)

		userID := "user_01jh6b8a2p6k43d4f4m0v4q2rs"
		tenantID := "tenant_01jh6b8a2p6k43d4f4m0v4q2rt"
		sessionID := "session_01jh6b8a2p6k43d4f4m0v4q2ru"
		duration := 5 * time.Minute

		token, err := maker.CreateAccessToken(userID, tenantID, sessionID, duration)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		claims, err := maker.VerifyAccessToken(token)
		require.NoError(t, err)
		require.NotNil(t, claims)
		require.Equal(t, userID, claims.UserID)
		require.Equal(t, tenantID, claims.TenantID)
		require.Equal(t, sessionID, claims.SessionID)
	})

	t.Run("Verify Token - expired", func(t *testing.T) {
		maker, err := NewJWTMaker(secretKey)
		require.NoError(t, err)

		userID := "user_123"
		tenantID := "tenant_123"
		sessionID := "session_123"
		duration := -5 * time.Minute // Expired in the past

		token, err := maker.CreateAccessToken(userID, tenantID, sessionID, duration)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		claims, err := maker.VerifyAccessToken(token)
		require.Error(t, err)
		require.Nil(t, claims)
		require.True(t, errors.Is(err, jwt.ErrTokenExpired))
	})

	t.Run("Verify Token - invalid signature", func(t *testing.T) {
		maker, err := NewJWTMaker(secretKey)
		require.NoError(t, err)

		userID := "user_123"
		tenantID := "tenant_123"
		sessionID := "session_123"
		duration := 5 * time.Minute

		token, err := maker.CreateAccessToken(userID, tenantID, sessionID, duration)
		require.NoError(t, err)

		otherSecretKey := "another_super_secret_key_minimum_32_characters"
		otherMaker, err := NewJWTMaker(otherSecretKey)
		require.NoError(t, err)

		claims, err := otherMaker.VerifyAccessToken(token)
		require.Error(t, err)
		require.Nil(t, claims)
		require.True(t, errors.Is(err, jwt.ErrSignatureInvalid))
	})

	t.Run("Verify Token - invalid format", func(t *testing.T) {
		maker, err := NewJWTMaker(secretKey)
		require.NoError(t, err)

		claims, err := maker.VerifyAccessToken("invalid-token-string")
		require.Error(t, err)
		require.Nil(t, claims)
	})

	t.Run("GenerateRefreshToken - success", func(t *testing.T) {
		maker, err := NewJWTMaker(secretKey)
		require.NoError(t, err)

		token1, err := maker.GenerateRefreshToken()
		require.NoError(t, err)
		require.NotEmpty(t, token1)

		token2, err := maker.GenerateRefreshToken()
		require.NoError(t, err)
		require.NotEmpty(t, token2)

		require.NotEqual(t, token1, token2)
	})
}
