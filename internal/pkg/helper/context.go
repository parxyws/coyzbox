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
