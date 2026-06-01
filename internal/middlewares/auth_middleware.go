package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/core"
	"github.com/parxyws/cozybox/pkg/helper"
)

// AuthMiddleware validates the JWT access token from the Authorization header
// and injects user_id, tenant_id, and session_id into the request context.
func (m *ManagerMiddleware) AuthMiddleware(jwt helper.TokenGenerator) gin.HandlerFunc {
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

		claims, err := jwt.VerifyAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Inject claims into Gin context (for handlers)
		c.Set(string(core.UserID), claims.UserID)
		c.Set(string(core.TenantID), claims.TenantID)
		c.Set(string(core.SessionID), claims.SessionID)

		// Inject into Go context (for service/repo layer)
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, core.UserID, claims.UserID)
		ctx = context.WithValue(ctx, core.TenantID, claims.TenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequirePasswordChangeMiddleware blocks access to protected routes if the user needs to change their password,
// except for the update-password route itself.
func (m *ManagerMiddleware) RequirePasswordChangeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for the password update route
		if c.Request.URL.Path == "/api/user/password" && c.Request.Method == http.MethodPut {
			c.Next()
			return
		}

		userID, exists := c.Get(string(core.UserID))
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var forcePasswordChange bool
		err := m.db.Table("users").Select("force_password_change").Where("id = ? AND deleted_at IS NULL", userID).Scan(&forcePasswordChange).Error
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to verify user status"})
			return
		}

		if forcePasswordChange {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "PASSWORD_CHANGE_REQUIRED", "message": "You must change your temporary password before accessing the platform."})
			return
		}

		c.Next()
	}
}
