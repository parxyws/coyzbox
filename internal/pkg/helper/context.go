package helper

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

type ContextKey string

const (
	RequestID ContextKey = "request_id"
	UserID    ContextKey = "user_id"
)

// DefaultTimeout provides a logical baseline timeout for operations
const DefaultTimeout = 10 * time.Second

// GetContext creates a context with timeout and request ID for service layer propagation
func GetContext(c *gin.Context) (context.Context, context.CancelFunc) {
	return GetContextWithTimeout(c, DefaultTimeout)
}

// GetContextWithTimeout creates a context with a custom timeout and request ID
func GetContextWithTimeout(c *gin.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)

	reqID := c.GetString("X-Request-ID")
	if reqID == "" {
		reqID = c.GetHeader("X-Request-ID")
	}

	ctx = context.WithValue(ctx, RequestID, reqID)
	return ctx, cancel
}
