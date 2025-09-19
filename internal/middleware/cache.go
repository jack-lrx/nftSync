package middleware

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

var (
	RedisClient *redis.Client
)

type Cache struct {
	redis *redis.Client
}

func NewRedis(redis *redis.Client) *Cache {
	return &Cache{
		redis: redis,
	}
}

func (c *Cache) SetCache(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.redis.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) GetCache(ctx context.Context, key string) (string, error) {
	return c.redis.Get(ctx, key).Result()
}

func (c *Cache) DelCache(ctx context.Context, key string) error {
	return c.redis.Del(ctx, key).Err()
}
