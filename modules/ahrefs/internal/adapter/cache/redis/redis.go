package redis

import (
	"context"

	"github.com/go-redis/redis/v9"
	"github.com/wahyudibo/golang-reverse-proxy/modules/ahrefs/internal/config"
)

var Prefix = "ahx"

func New(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.CacheRedisAddress,
		Password: cfg.CacheRedisPassword,
		DB:       0,
	})

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return client, nil
}
