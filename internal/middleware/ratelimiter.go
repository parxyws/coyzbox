package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	redisclient "github.com/redis/go-redis/v9"
)

// RateLimitConfig defines the configuration for a rate limiter.
type RateLimitConfig struct {
	MaxRequests int                         // maximum requests allowed in the window
	Window      time.Duration               // time window
	KeyFunc     func(c *gin.Context) string // key generator (e.g. by IP)
}

// NewRateLimiter returns a Gin middleware that enforces a fixed-window
// rate limit using a Redis counter with TTL. On Redis failure it fails
// open (allows the request through).
func NewRateLimiter(redisClient *redisclient.Client, cfg RateLimitConfig) gin.HandlerFunc {
	windowSec := int(cfg.Window.Seconds())

	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := cfg.KeyFunc(c)

		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			redisClient.Expire(ctx, key, cfg.Window)
		}

		remaining := cfg.MaxRequests - int(count)
		if remaining < 0 {
			remaining = 0
		}

		ttl, err := redisClient.TTL(ctx, key).Result()
		if err != nil {
			ttl = cfg.Window
		}
		resetAfter := int(ttl.Seconds())
		if resetAfter < 0 {
			resetAfter = windowSec
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.MaxRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.Itoa(resetAfter))

		if count > int64(cfg.MaxRequests) {
			c.Header("Retry-After", strconv.Itoa(resetAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "too many requests",
				"retry_after": resetAfter,
			})
			return
		}

		c.Next()
	}
}

// RateLimitByIP returns a key function scoped to client IP.
func RateLimitByIP(prefix string) func(c *gin.Context) string {
	return func(c *gin.Context) string {
		return fmt.Sprintf("ratelimit:%s:%s", prefix, c.ClientIP())
	}
}
