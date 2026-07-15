package helper

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/parxyws/cozybox/internal/config"
)

const DefaultTimeout = 10 * time.Second

func GetContext(c *gin.Context) (context.Context, context.CancelFunc) {
	return GetContextWithTimeout(c, DefaultTimeout)
}

func GetContextWithTimeout(c *gin.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)

	reqID := c.GetString("X-Request-ID")
	if reqID == "" {
		reqID = c.GetHeader("X-Request-ID")
	}

	ctx = context.WithValue(ctx, config.RequestID, reqID)
	return ctx, cancel
}

// UserIDFromContext extracts the authenticated user's ID from the context.
// Returns the ID and true on success; empty string and false if missing.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(config.UserID).(string)
	return id, ok && id != ""
}

// TenantIDFromContext extracts the active tenant ID from the context.
// Returns the ID and true on success; empty string and false if missing.
func TenantIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(config.TenantID).(string)
	return id, ok && id != ""
}

// SessionIDFromContext extracts the session ID from the context.
// Returns the ID and true on success; empty string and false if missing.
func SessionIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(config.SessionID).(string)
	return id, ok && id != ""
}

// RequestIDFromContext extracts the request trace ID from the context.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(config.RequestID).(string)
	return id
}
