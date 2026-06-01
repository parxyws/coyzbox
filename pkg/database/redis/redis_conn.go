package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/parxyws/cozybox/internal/config"
	redisclient "github.com/redis/go-redis/v9"
)

// newRedisClient creates a Redis client with production-grade settings.
// All pool, timeout, and retry options follow go-redis best practices.
func newRedisClient(addr, password string, db int) (*redisclient.Client, error) {
	client := redisclient.NewClient(&redisclient.Options{
		Addr:     addr,
		Password: password,
		DB:       db,

		// Connection pool
		PoolSize:       100,
		MinIdleConns:   10,
		MaxIdleConns:   20,
		MaxActiveConns: 100,

		// Timeouts
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,

		// Connection lifetime
		ConnMaxIdleTime: 30 * time.Minute,

		// Retry settings
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis db %d: %w", db, err)
	}

	return client, nil
}

// InitAuthRedis initializes the Redis client for authentication
func InitAuthRedis(cfg *config.Config) (*redisclient.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	return newRedisClient(addr, cfg.Redis.Password, cfg.Redis.AuthDB)
}

// InitCacheRedis initializes the Redis client for caching
func InitCacheRedis(cfg *config.Config) (*redisclient.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	return newRedisClient(addr, cfg.Redis.Password, cfg.Redis.CacheDB)
}

// InitLimiterRedis initializes the Redis client for rate limiting
func InitLimiterRedis(cfg *config.Config) (*redisclient.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	return newRedisClient(addr, cfg.Redis.Password, cfg.Redis.LimiterDB)
}
