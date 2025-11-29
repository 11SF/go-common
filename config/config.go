package commonconfig

import (
	"context"

	"github.com/11SF/go-common/logger"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const (
	tag = "load config func"
)

func LoadConfig[CONFIG any](ctx context.Context, envPath string) (*CONFIG, error) {
	logger.Info(ctx, "loading environment variables", logger.LogAttrTag(tag))
	err := godotenv.Load(envPath)
	if err != nil {
		logger.Error(ctx, "error loading .env file", logger.LogAttrError(err), logger.LogAttrTag(tag))
		return nil, err
	}

	cfg, err := env.ParseAs[CONFIG]()
	if err != nil {
		logger.Error(ctx, "error loading environment variables", logger.LogAttrError(err), logger.LogAttrTag(tag))
		return nil, err
	}

	logger.Info(ctx, "environment variables loaded", logger.LogAttrTag(tag))
	return &cfg, nil
}
