package logger

import (
	"time"

	"github.com/11SF/go-common/telemetry"

	"github.com/gin-gonic/gin"
)

func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		ctx := c.Request.Context()

		// Get trace/span IDs from OpenTelemetry context (set by telemetry middleware)
		traceID := telemetry.TraceID(ctx)
		spanID := telemetry.SpanID(ctx)

		// Fallback to custom tracing if OpenTelemetry is not available
		if traceID == "" || spanID == "" {
			ctx = NewTraceContext(ctx)
			c.Request = c.Request.WithContext(ctx)
			if traceID == "" {
				traceID = GetTraceID(ctx)
			}
			if spanID == "" {
				spanID = GetSpanID(ctx)
			}
		}

		c.Header("X-Trace-ID", traceID)
		c.Header("X-Span-ID", spanID)

		c.Next()

		timestamp := time.Now()
		latency := timestamp.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()
		userAgent := c.Request.UserAgent()

		if raw != "" {
			path = path + "?" + raw
		}

		if statusCode >= 400 && statusCode < 500 {
			Warn(ctx, "HTTP Request",
				"method", method,
				"path", path,
				"status", statusCode,
				"latency", latency.String(),
				"client_ip", clientIP,
				"user_agent", userAgent,
				"body_size", bodySize,
			)
		} else if statusCode >= 500 {
			Error(ctx, "HTTP Request",
				"method", method,
				"path", path,
				"status", statusCode,
				"latency", latency.String(),
				"client_ip", clientIP,
				"user_agent", userAgent,
				"body_size", bodySize,
			)
		} else {
			Info(ctx, "HTTP Request",
				"method", method,
				"path", path,
				"status", statusCode,
				"latency", latency.String(),
				"client_ip", clientIP,
				"user_agent", userAgent,
				"body_size", bodySize,
			)
		}
	}
}

func GinLogger(logLevel string) *Logger {
	return New(getLogLevel(logLevel))
}
