package middleware

import (
	"log/slog"
	"time"
)

// Config defines the configuration for framework middleware logging.
type Config struct {
	// Logger is the slog.Logger to use. If nil, slog.Default() is used.
	Logger *slog.Logger

	// LogRequest enables logging of incoming requests (method, path, query).
	LogRequest bool

	// LogResponse enables logging of outgoing responses (status, latency).
	LogResponse bool

	// LogError enables logging of errors returned by handlers.
	LogError bool

	// LogRequestData enables logging of request body content.
	// WARNING: May log sensitive data. Use with caution.
	LogRequestData bool

	// LogResponseData enables logging of response body content.
	// WARNING: May log sensitive data. Use with caution.
	LogResponseData bool

	// LogHeaders enables logging of specific request/response headers.
	LogHeaders bool

	// Headers is the list of headers to log when LogHeaders is true.
	// If empty and LogHeaders is true, all headers are logged.
	Headers []string

	// TraceIDHeader is the HTTP header to extract a trace ID from.
	// If set, the trace ID is included in all log entries for the request.
	TraceIDHeader string

	// SkipPaths is a list of request paths that should not be logged.
	SkipPaths []string

	// MaxBodySize limits the body size (in bytes) logged when body logging is enabled.
	MaxBodySize int

	// TimeFormat specifies the time format for log timestamps.
	TimeFormat string

	// SkipFunc allows custom logic to skip logging for specific requests.
	SkipFunc func(method, path string, status int) bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		LogRequest:      true,
		LogResponse:     true,
		LogError:        true,
		LogRequestData:  false,
		LogResponseData: false,
		LogHeaders:      false,
		TraceIDHeader:   "",
		SkipPaths:       []string{"/health", "/ready", "/metrics"},
		MaxBodySize:     10 * 1024,
		TimeFormat:      time.RFC3339,
	}
}
