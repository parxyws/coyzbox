package redis

import (
	"context"
	"fmt"
	"time"

	redisclient "github.com/redis/go-redis/v9"
)

type SessionStore struct{ client *redisclient.Client }

func NewSessionStore(client *redisclient.Client) *SessionStore { return &SessionStore{client} }

func (s *SessionStore) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return s.client.Set(ctx, key, value, expiration).Err()
}

func (s *SessionStore) SetMultiple(ctx context.Context, pairs map[string]string, expiration time.Duration) error {
	pipe := s.client.Pipeline()
	for k, v := range pairs {
		pipe.Set(ctx, k, v, expiration)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("pipeline exec: %w", err)
	}
	return nil
}

func (s *SessionStore) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

func (s *SessionStore) Delete(ctx context.Context, keys ...string) error {
	return s.client.Del(ctx, keys...).Err()
}
