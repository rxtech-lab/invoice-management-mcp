package services

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisService provides caching functionality using Redis
type RedisService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Close() error
}

type redisService struct {
	client *redis.Client
}

// NewRedisService creates a new Redis service instance
// Returns nil if connection fails
func NewRedisService(redisURL string) (RedisService, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &redisService{client: client}, nil
}

func (r *redisService) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisService) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisService) Close() error {
	return r.client.Close()
}
