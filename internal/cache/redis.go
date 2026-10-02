package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"taskhub/internal/configs"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps redis.Client with helper methods
type RedisClient struct {
	client *redis.Client
}

// InitRedis connects to Redis and tests the connection
func InitRedis(cfg *configs.Config) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &RedisClient{client: rdb}, nil
}

// Client returns underlying go-redis client
func (r *RedisClient) Client() *redis.Client {
	if r == nil {
		return nil
	}
	return r.client
}

// Ping checks if Redis is reachable
func (r *RedisClient) Ping(ctx context.Context) error {
	if r == nil || r.client == nil {
		return fmt.Errorf("redis client is nil")
	}
	return r.client.Ping(ctx).Err()
}

// Get retrieves and unmarshals JSON value from Redis
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	if r == nil || r.client == nil {
		return redis.Nil
	}

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// Set marshals value to JSON and stores in Redis with TTL
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if r == nil || r.client == nil {
		return nil
	}

	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	return r.client.Set(ctx, key, bytes, ttl).Err()
}

// Delete removes a specific key from Redis
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Del(ctx, key).Err()
}

// DeleteByPattern finds and deletes all keys matching glob pattern
func (r *RedisClient) DeleteByPattern(ctx context.Context, pattern string) error {
	if r == nil || r.client == nil {
		return nil
	}

	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}
