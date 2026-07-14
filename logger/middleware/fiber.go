package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
)

// FiberMiddleware returns a Fiber middleware that auto-logs requests and responses.
func FiberMiddleware(cfg Config) fiber.Handler {
	logger := slog.Default()

	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}

	return func(c *fiber.Ctx) error {
		// Check if path should be skipped
		reqPath := c.Path()
		for _, skip := range cfg.SkipPaths {
			if reqPath == skip {
				return c.Next()
			}
		}
		if cfg.SkipFunc != nil && cfg.SkipFunc(c.Method(), reqPath, 0) {
			return c.Next()
		}

		start := time.Now()

		// Trace ID extraction
		traceID := ""
		if cfg.TraceIDHeader != "" {
			traceID = c.Get(cfg.TraceIDHeader)
		}

		// ---- Log incoming request ----
		if cfg.LogRequest {
			attrs := []slog.Attr{
				slog.String("method", c.Method()),
				slog.String("path", reqPath),
				slog.String("query", string(c.Request().URI().QueryString())),
				slog.String("remote_addr", c.IP()),
				slog.String("user_agent", c.Get("User-Agent")),
			}
			if traceID != "" {
				attrs = append(attrs, slog.String("trace_id", traceID))
			}

			if cfg.LogRequestData {
				bodyBytes := make([]byte, len(c.Body()))
				copy(bodyBytes, c.Body())
				bodyStr := string(bodyBytes)
				if cfg.MaxBodySize > 0 && len(bodyStr) > cfg.MaxBodySize {
					bodyStr = bodyStr[:cfg.MaxBodySize] + "..."
				}
				attrs = append(attrs, slog.String("body", bodyStr))
			}

			logger.LogAttrs(c.UserContext(), slog.LevelInfo, "incoming request", attrs...)
		}

		// Process request
		if err := c.Next(); err != nil {
			return err
		}

		// ---- Log response ----
		latency := time.Since(start)
		status := c.Response().StatusCode()

		if cfg.LogResponse {
			attrs := []slog.Attr{
				slog.Int("status", status),
				slog.String("latency", latency.String()),
				slog.Int64("latency_ms", latency.Milliseconds()),
			}
			if traceID != "" {
				attrs = append(attrs, slog.String("trace_id", traceID))
			}

			if cfg.LogResponseData {
				bodyStr := string(c.Response().Body())
				if cfg.MaxBodySize > 0 && len(bodyStr) > cfg.MaxBodySize {
					bodyStr = bodyStr[:cfg.MaxBodySize] + "..."
				}
				attrs = append(attrs, slog.String("body", bodyStr))
			}

			logger.LogAttrs(c.UserContext(), slog.LevelInfo, "outgoing response", attrs...)
		}

		return nil
	}
}
