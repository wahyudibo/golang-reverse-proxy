package redis

import (
	"github.com/go-redis/redis/v9"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/config"
)

func New(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.CacheRedisAddress,
		Password: cfg.CacheRedisPassword,
		DB:       0,
	})
}
