package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"github.com/parxyws/cozybox/internal/config"
)

const headerRequestID = "X-Request-ID"

// RequestID is a middleware that reads or generates a unique request ID.
// It sets the ID in both the Gin context and the Go context for downstream use,
// and writes it to the response header for client-side traceability.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(headerRequestID)
		if requestID == "" {
			requestID = ulid.Make().String()
		}

		// Set in Gin context
		c.Set(string(config.RequestID), requestID)

		// Set in Go context for downstream services
		ctx := context.WithValue(c.Request.Context(), config.RequestID, requestID)
		c.Request = c.Request.WithContext(ctx)

		// Write to response header
		c.Header(headerRequestID, requestID)

		c.Next()
	}
}
