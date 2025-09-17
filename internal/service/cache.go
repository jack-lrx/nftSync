package service

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

var (
	RedisClient *redis.Client
)

func InitRedis(addr, password string, db int) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

func SetCache(ctx context.Context, key string, value string, ttl time.Duration) error {
	return RedisClient.Set(ctx, key, value, ttl).Err()
}

func GetCache(ctx context.Context, key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}

func DelCache(ctx context.Context, key string) error {
	return RedisClient.Del(ctx, key).Err()
}
