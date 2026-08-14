package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/parxyws/cozybox/internal/config"
	"github.com/redis/go-redis/v9"
)

func InitClient(cfg *config.Config, db int) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
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
		if closeErr := client.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to connect to redis db %d: %w (close error: %v)", db, err, closeErr)
		}
		return nil, fmt.Errorf("failed to connect to redis db %d: %w", db, err)
	}

	return client, nil
}
