package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	redisclient "github.com/redis/go-redis/v9"
)

// RateLimitConfig defines the configuration for the rate limiter.
type RateLimitConfig struct {
	MaxRequests int                         // Maximum number of requests allowed in the window
	Window      time.Duration               // Time window for the rate limit
	KeyFunc     func(c *gin.Context) string // Function to generate the rate limit key
}

// NewRateLimiter returns a Gin middleware that enforces a sliding-window rate limit
// using Redis sorted sets. When the limit is exceeded, it returns 429 Too Many Requests.
func NewRateLimiter(redisClient *redisclient.Client, cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := cfg.KeyFunc(c)

		now := time.Now()
		windowStart := now.Add(-cfg.Window)

		pipe := redisClient.Pipeline()

		// Remove entries outside the window
		pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.UnixMilli(), 10))

		// Count current entries in the window
		countCmd := pipe.ZCard(ctx, key)

		// Add the current request
		member := fmt.Sprintf("%s:%d", ulid.Make().String(), now.UnixNano())
		pipe.ZAdd(ctx, key, redisclient.Z{
			Score:  float64(now.UnixMilli()),
			Member: member,
		})

		// Set expiry on the key to auto-cleanup
		pipe.Expire(ctx, key, cfg.Window+time.Second)

		if _, err := pipe.Exec(ctx); err != nil {
			// On Redis failure, allow the request (fail-open)
			c.Next()
			return
		}

		count := countCmd.Val()
		if count >= int64(cfg.MaxRequests) {
			retryAfter := int(cfg.Window.Seconds())
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.MaxRequests))
			c.Header("X-RateLimit-Remaining", "0")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "too many requests",
				"retry_after": retryAfter,
			})
			return
		}

		// Set rate limit headers
		remaining := int64(cfg.MaxRequests) - count - 1
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.MaxRequests))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

		c.Next()
	}
}

// RateLimitByIP returns a KeyFunc that generates rate limit keys based on client IP.
func RateLimitByIP(prefix string) func(c *gin.Context) string {
	return func(c *gin.Context) string {
		return fmt.Sprintf("ratelimit:%s:%s", prefix, c.ClientIP())
	}
}

// RateLimitByBodyField returns a KeyFunc that generates rate limit keys based on
// a JSON body field (e.g., "email"). Falls back to client IP if the field is absent.
func RateLimitByBodyField(prefix string, field string) func(c *gin.Context) string {
	return func(c *gin.Context) string {
		val := c.PostForm(field)
		if val == "" {
			// Try to peek at JSON body without consuming it
			val = c.ClientIP()
		}
		return fmt.Sprintf("ratelimit:%s:%s", prefix, val)
	}
}
