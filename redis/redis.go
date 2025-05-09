package redisclient

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
}

func NewRedisClient(cfg Config) redis.UniversalClient {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
	})

	err := redisClient.Ping(context.Background()).Err()
	if err != nil {
		panic(err)
	}

	return redisClient
}
