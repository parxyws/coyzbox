package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/config"
	"github.com/parxyws/cozybox/internal/domain"
	"github.com/parxyws/cozybox/internal/pkg/jwtutil"
)

// SessionValidator retrieves session data from the session store.
// It mirrors the Get method from the SessionStore interface, kept minimal
// to avoid importing the full auth package.
type SessionValidator interface {
	Get(ctx context.Context, key string) (string, error)
}

// NewAuthMiddleware creates a Gin middleware that authenticates requests via JWT
// and validates that the session is still active in Redis.
//
// Flow:
//  1. Extract Bearer token from Authorization header
//  2. Verify JWT signature and claims
//  3. Validate that the session still exists in Redis (immediate logout enforcement)
//  4. Verify the user_id in the session matches the JWT claim (session hijacking detection)
//  5. Set user_id, tenant_id, session_id in both Gin and Go context
func NewAuthMiddleware(tokenGen jwtutil.TokenGenerator, sessionStore SessionValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		claims, err := tokenGen.VerifyAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Validate session is still active in Redis.
		// This ensures that after logout (session deleted), the access token
		// is immediately rejected rather than remaining valid until JWT expiry.
		sessionJSON, err := sessionStore.Get(c.Request.Context(), fmt.Sprintf("session:%s", claims.SessionID))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired or invalidated"})
			return
		}

		// Verify the session's user_id matches the JWT claim to detect session hijacking.
		var session domain.UserSession
		if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid session data"})
			return
		}

		if session.UserID != claims.UserID {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session mismatch"})
			return
		}

		// Check if session has expired (belt-and-suspenders with Redis TTL)
		if time.Now().After(session.ExpiresAt) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
			return
		}

		c.Set(string(config.UserID), claims.UserID)
		c.Set(string(config.TenantID), claims.TenantID)
		c.Set(string(config.SessionID), claims.SessionID)

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, config.UserID, claims.UserID)
		ctx = context.WithValue(ctx, config.TenantID, claims.TenantID)
		ctx = context.WithValue(ctx, config.SessionID, claims.SessionID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
