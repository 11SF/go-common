package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// responseWriter wraps gin.ResponseWriter to capture the status code and body.
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) WriteString(s string) (int, error) {
	rw.body.WriteString(s)
	return rw.ResponseWriter.WriteString(s)
}

// GinMiddleware returns a Gin handler that auto-logs requests and responses.
func GinMiddleware(cfg Config) gin.HandlerFunc {
	logger := slog.Default()

	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}

	return func(c *gin.Context) {
		// Check if path should be skipped
		reqPath := c.Request.URL.Path
		for _, skip := range cfg.SkipPaths {
			if reqPath == skip {
				c.Next()
				return
			}
		}
		if cfg.SkipFunc != nil && cfg.SkipFunc(c.Request.Method, reqPath, 0) {
			c.Next()
			return
		}

		start := time.Now()

		// Trace ID extraction
		traceID := ""
		if cfg.TraceIDHeader != "" {
			traceID = c.Request.Header.Get(cfg.TraceIDHeader)
		}

		// ---- Log incoming request ----
		if cfg.LogRequest {
			attrs := []slog.Attr{
				slog.String("method", c.Request.Method),
				slog.String("path", reqPath),
				slog.String("query", c.Request.URL.RawQuery),
				slog.String("remote_addr", c.ClientIP()),
				slog.String("user_agent", c.Request.UserAgent()),
			}
			if traceID != "" {
				attrs = append(attrs, slog.String("trace_id", traceID))
			}

			if cfg.LogHeaders && len(cfg.Headers) > 0 {
				headerAttrs := make([]slog.Attr, 0, len(cfg.Headers))
				for _, h := range cfg.Headers {
					if v := c.Request.Header.Get(h); v != "" {
						headerAttrs = append(headerAttrs, slog.String(h, v))
					}
				}
				if len(headerAttrs) > 0 {
					attrs = append(attrs, slog.Any("headers", slog.GroupValue(headerAttrs...)))
				}
			} else if cfg.LogHeaders {
				attrs = append(attrs, slog.Any("headers", c.Request.Header))
			}

			if cfg.LogRequestData {
				bodyBytes, _ := io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				bodyStr := string(bodyBytes)
				if cfg.MaxBodySize > 0 && len(bodyStr) > cfg.MaxBodySize {
					bodyStr = bodyStr[:cfg.MaxBodySize] + "..."
				}
				attrs = append(attrs, slog.String("body", bodyStr))
			}

			logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "incoming request", attrs...)
		}

		// Wrap response writer if we need to capture response
		if cfg.LogRequestData || cfg.LogResponseData || cfg.LogResponse {
			rw := &responseWriter{
				ResponseWriter: c.Writer,
				body:           &bytes.Buffer{},
			}
			c.Writer = rw
		}

		// Process request
		c.Next()

		// ---- Log response ----
		latency := time.Since(start)
		status := c.Writer.Status()

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
				if rw, ok := c.Writer.(*responseWriter); ok {
					bodyStr := rw.body.String()
					if cfg.MaxBodySize > 0 && len(bodyStr) > cfg.MaxBodySize {
						bodyStr = bodyStr[:cfg.MaxBodySize] + "..."
					}
					attrs = append(attrs, slog.String("body", bodyStr))
				}
			}

			logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "outgoing response", attrs...)
		}

		// ---- Log errors ----
		if cfg.LogError && len(c.Errors) > 0 {
			for _, err := range c.Errors {
				stack := make([]byte, 4096)
				n := runtime.Stack(stack, false)
				logger.LogAttrs(c.Request.Context(), slog.LevelError, "handler error",
					slog.String("error", err.Err.Error()),
					slog.Int("status", status),
					slog.String("method", c.Request.Method),
					slog.String("path", reqPath),
					slog.String("stack", string(stack[:n])),
				)
			}
		}
	}
}
