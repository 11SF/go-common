package logger

import (
	"log/slog"

	"github.com/11SF/go-common/logger/middleware"
	"github.com/gin-gonic/gin"
)

// GinMiddleware returns a Gin middleware that auto-logs HTTP requests.
// Uses default config (skips /health, /ready, /metrics).
// For custom config, use middleware.GinMiddleware(cfg) directly.
func GinMiddleware() gin.HandlerFunc {
	cfg := middleware.DefaultConfig()
	cfg.Logger = slog.Default()
	return middleware.GinMiddleware(cfg)
}
