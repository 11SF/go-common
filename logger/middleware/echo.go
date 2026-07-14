package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

// EchoMiddleware returns an Echo middleware that auto-logs requests and responses.
func EchoMiddleware(cfg Config) echo.MiddlewareFunc {
	logger := slog.Default()

	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqPath := c.Path()
			for _, skip := range cfg.SkipPaths {
				if reqPath == skip {
					return next(c)
				}
			}
			if cfg.SkipFunc != nil && cfg.SkipFunc(c.Request().Method, reqPath, 0) {
				return next(c)
			}

			start := time.Now()

			// Trace ID extraction
			traceID := ""
			if cfg.TraceIDHeader != "" {
				traceID = c.Request().Header.Get(cfg.TraceIDHeader)
			}

			// ---- Log incoming request ----
			if cfg.LogRequest {
				attrs := []slog.Attr{
					slog.String("method", c.Request().Method),
					slog.String("path", reqPath),
					slog.String("query", c.Request().URL.RawQuery),
					slog.String("remote_addr", c.Request().RemoteAddr),
					slog.String("user_agent", c.Request().UserAgent()),
				}
				if traceID != "" {
					attrs = append(attrs, slog.String("trace_id", traceID))
				}

				if cfg.LogRequestData {
					bodyBytes, _ := io.ReadAll(c.Request().Body)
					c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					bodyStr := string(bodyBytes)
					if cfg.MaxBodySize > 0 && len(bodyStr) > cfg.MaxBodySize {
						bodyStr = bodyStr[:cfg.MaxBodySize] + "..."
					}
					attrs = append(attrs, slog.String("body", bodyStr))
				}

				logger.LogAttrs(c.Request().Context(), slog.LevelInfo, "incoming request", attrs...)
			}

			// Process request
			if err := next(c); err != nil {
				c.Error(err)
			}

			// ---- Log response ----
			latency := time.Since(start)
			res := c.Response()
			status := res.Status

			if cfg.LogResponse {
				attrs := []slog.Attr{
					slog.Int("status", status),
					slog.String("latency", latency.String()),
					slog.Int64("latency_ms", latency.Milliseconds()),
				}
				if traceID != "" {
					attrs = append(attrs, slog.String("trace_id", traceID))
				}

				logger.LogAttrs(c.Request().Context(), slog.LevelInfo, "outgoing response", attrs...)
			}

			return nil
		}
	}
}
