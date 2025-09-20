package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"

	"github.com/11SF/go-common/telemetry"
)

type contextKey string

const (
	TraceIDKey contextKey = "trace_id"
	SpanIDKey  contextKey = "span_id"
)

type Logger struct {
	*slog.Logger
}

func New(logLevel slog.Leveler) *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Key = "timestamp"
			}
			if a.Key == slog.LevelKey {
				a.Key = "level"
			}
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}
			return a
		},
	})

	return &Logger{
		Logger: slog.New(handler),
	}
}

func Init(logLevel string) {
	logger := New(getLogLevel(logLevel))
	slog.SetDefault(logger.Logger)
}

func getLogLevel(logLevel string) slog.Leveler {
	var slogLevel slog.Leveler
	switch logLevel {
	case "INFO":
		slogLevel = slog.LevelInfo
	case "ERROR":
		slogLevel = slog.LevelError
	case "DEBUG":
		slogLevel = slog.LevelDebug
	case "WARN":
		slogLevel = slog.LevelWarn
	default:
		slogLevel = slog.LevelInfo
	}

	return slogLevel
}

func (l *Logger) WithTracing(ctx context.Context) *slog.Logger {
	var args []any

	// Try OpenTelemetry first
	if traceID := telemetry.TraceID(ctx); traceID != "" {
		args = append(args, "trace_id", traceID)
	} else if traceID := GetTraceID(ctx); traceID != "" {
		args = append(args, "trace_id", traceID)
	}

	if spanID := telemetry.SpanID(ctx); spanID != "" {
		args = append(args, "span_id", spanID)
	} else if spanID := GetSpanID(ctx); spanID != "" {
		args = append(args, "span_id", spanID)
	}

	return l.Logger.With(args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	logger := slog.Default()
	if l, ok := logger.Handler().(*slog.JSONHandler); ok {
		tempLogger := &Logger{Logger: slog.New(l)}
		tempLogger.WithTracing(ctx).Info(msg, args...)
		return
	}
	logger.InfoContext(ctx, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	logger := slog.Default()
	if l, ok := logger.Handler().(*slog.JSONHandler); ok {
		tempLogger := &Logger{Logger: slog.New(l)}
		tempLogger.WithTracing(ctx).Error(msg, args...)
		return
	}
	logger.ErrorContext(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	logger := slog.Default()
	if l, ok := logger.Handler().(*slog.JSONHandler); ok {
		tempLogger := &Logger{Logger: slog.New(l)}
		tempLogger.WithTracing(ctx).Warn(msg, args...)
		return
	}
	logger.WarnContext(ctx, msg, args...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	logger := slog.Default()
	if l, ok := logger.Handler().(*slog.JSONHandler); ok {
		tempLogger := &Logger{Logger: slog.New(l)}
		tempLogger.WithTracing(ctx).Debug(msg, args...)
		return
	}
	logger.DebugContext(ctx, msg, args...)
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, SpanIDKey, spanID)
}

func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

func GetSpanID(ctx context.Context) string {
	if spanID, ok := ctx.Value(SpanIDKey).(string); ok {
		return spanID
	}
	return ""
}

func GenerateTraceID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func GenerateSpanID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func NewTraceContext(ctx context.Context) context.Context {
	traceID := GenerateTraceID()
	spanID := GenerateSpanID()

	ctx = WithTraceID(ctx, traceID)
	ctx = WithSpanID(ctx, spanID)

	return ctx
}

func LogAttrError(err error) slog.Attr {
	return slog.String("err", err.Error())
}

func LogAttrTag(tag string) slog.Attr {
	return slog.String("tag", tag)
}
