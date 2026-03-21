package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"moneyapp/backend/internal/config"

	goredis "github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *goredis.Client
}

func NewRedisStore(ctx context.Context, cfg config.RedisConfig) (*RedisStore, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &RedisStore{client: client}, nil
}

func (s *RedisStore) GetJSON(ctx context.Context, key string, dst any) (bool, error) {
	value, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, goredis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if err := json.Unmarshal([]byte(value), dst); err != nil {
		return false, err
	}

	return true, nil
}

func (s *RedisStore) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, key, payload, ttl).Err()
}

func (s *RedisStore) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	return s.client.Del(ctx, keys...).Err()
}

func (s *RedisStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}
