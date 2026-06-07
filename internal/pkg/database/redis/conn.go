package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/parxyws/cozybox/internal/config"
	redisclient "github.com/redis/go-redis/v9"
)

type Client struct {
	*redisclient.Client
}

func newRedisClient(addr, password string, db int) (*Client, error) {
	client := redisclient.NewClient(&redisclient.Options{
		Addr:     addr,
		Password: password,
		DB:       db,

		PoolSize:       100,
		MinIdleConns:   10,
		MaxIdleConns:   20,
		MaxActiveConns: 100,

		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,

		ConnMaxIdleTime: 30 * time.Minute,

		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis db %d: %w", db, err)
	}

	return &Client{Client: client}, nil
}

func InitAuthRedis(cfg *config.Config) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	return newRedisClient(addr, cfg.Redis.Password, cfg.Redis.AuthDB)
}

func InitCacheRedis(cfg *config.Config) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	return newRedisClient(addr, cfg.Redis.Password, cfg.Redis.CacheDB)
}

func InitLimiterRedis(cfg *config.Config) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	return newRedisClient(addr, cfg.Redis.Password, cfg.Redis.LimiterDB)
}
