package redisclient

import (
	"context"
	"log/slog"

	"github.com/11SF/go-common/logger"
	"github.com/redis/go-redis/v9"
)

const (
	tag = "redis client"
)

type Config struct {
	Addr     string `env:"REDIS_ADDRESS"`
	Password string `env:"REDIS_PASSWORD"`
}

func NewRedisClient(ctx context.Context, cfg Config) redis.UniversalClient {

	logger.Info(ctx, "connecting to redis", slog.String("addr", cfg.Addr), logger.LogAttrTag(tag))
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
	})

	err := redisClient.Ping(ctx).Err()
	if err != nil {
		logger.Error(ctx, "fail to connect redis", logger.LogAttrError(err), logger.LogAttrTag(tag))
		panic(err)
	}

	logger.Info(ctx, "connect redis success", slog.String("addr", cfg.Addr), logger.LogAttrTag(tag))
	return redisClient
}
