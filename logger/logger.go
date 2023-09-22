package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	LogLevel string
}

func CreateLogger(cf Config) *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	zapLevel := zap.InfoLevel
	if cf.LogLevel == "debug" {
		zapLevel = zap.DebugLevel
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},

		InitialFields: map[string]interface{}{
			"pid":      os.Getpid(),
			"hostname": hostname,
		},
	}

	return zap.Must(config.Build())
}
