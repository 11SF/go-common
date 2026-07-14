// Package logger provides a common logging wrapper around Go's log/slog
// with auto-logging middleware for Gin, Fiber, and Echo frameworks.
package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
)

// Logger wraps slog.Logger and provides convenience methods for structured logging.
type Logger struct {
	*slog.Logger
}

// New creates a new Logger wrapping the provided slog.Logger.
// If l is nil, it uses slog.Default().
func New(l *slog.Logger) *Logger {
	if l == nil {
		l = slog.Default()
	}
	return &Logger{Logger: l}
}

// NewJSONLogger creates a new Logger that outputs JSON to the given writer.
// If w is nil, output goes to os.Stdout.
func NewJSONLogger(w *os.File) *Logger {
	if w == nil {
		w = os.Stdout
	}
	return New(slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
}

// NewTextLogger creates a new Logger that outputs human-readable text to the given writer.
// If w is nil, output goes to os.Stdout.
func NewTextLogger(w *os.File) *Logger {
	if w == nil {
		w = os.Stdout
	}
	return New(slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
}

// With adds structured attributes to the logger.
// Usage: logger.With("service", "my-service", "version", "1.0.0")
func (l *Logger) With(args ...any) *Logger {
	return &Logger{Logger: l.Logger.With(args...)}
}

// Request logs an incoming request event.
func (l *Logger) Request(ctx context.Context, msg string, attrs ...any) {
	l.LogAttrs(ctx, slog.LevelInfo, msg, argsToAttrs(attrs)...)
}

// Response logs an outgoing response event.
func (l *Logger) Response(ctx context.Context, msg string, attrs ...any) {
	l.LogAttrs(ctx, slog.LevelInfo, msg, argsToAttrs(attrs)...)
}

// Error logs an error event with stack trace information.
func (l *Logger) Error(ctx context.Context, msg string, err error, attrs ...any) {
	if err == nil {
		l.LogAttrs(ctx, slog.LevelError, msg, argsToAttrs(attrs)...)
		return
	}

	// Capture stack trace
	stack := make([]byte, 4096)
	n := runtime.Stack(stack, false)
	stack = stack[:n]

	errorAttrs := append([]any{
		slog.String("error", err.Error()),
		slog.String("stack", string(stack)),
	}, attrs...)

	l.LogAttrs(ctx, slog.LevelError, msg, argsToAttrs(errorAttrs)...)
}

// Warn logs a warning event.
func (l *Logger) Warn(ctx context.Context, msg string, attrs ...any) {
	l.LogAttrs(ctx, slog.LevelWarn, msg, argsToAttrs(attrs)...)
}

// Info logs an info event.
func (l *Logger) Info(ctx context.Context, msg string, attrs ...any) {
	l.LogAttrs(ctx, slog.LevelInfo, msg, argsToAttrs(attrs)...)
}

// Debug logs a debug event.
func (l *Logger) Debug(ctx context.Context, msg string, attrs ...any) {
	l.LogAttrs(ctx, slog.LevelDebug, msg, argsToAttrs(attrs)...)
}

// LogAttrs logs a message with Level and slog.Attr values.
// This is a thin wrapper around slog.Logger.LogAttrs that preserves the Logger wrapper.
func (l *Logger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.Logger.LogAttrs(ctx, level, msg, attrs...)
}

// argsToAttrs converts a mixed slice of any values to slog.Attr values.
// Supports: slog.Attr, string key-value pairs, and error types.
func argsToAttrs(args []any) []slog.Attr {
	if len(args) == 0 {
		return nil
	}

	attrs := make([]slog.Attr, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch v := args[i].(type) {
		case slog.Attr:
			attrs = append(attrs, v)
		case string:
			// Treat as key-value pair: look ahead for the value
			if i+1 < len(args) {
				i++
				attrs = append(attrs, slog.Any(v, args[i]))
			} else {
				attrs = append(attrs, slog.String("key", v))
			}
		case error:
			attrs = append(attrs, slog.String("error", v.Error()))
		default:
			attrs = append(attrs, slog.Any("value", v))
		}
	}
	return attrs
}

// ─── Package-level convenience functions (use slog.Default()) ───

// Info logs at info level using the default logger.
func Info(ctx context.Context, msg string, args ...any) {
	slog.Default().LogAttrs(ctx, slog.LevelInfo, msg, argsToAttrs(args)...)
}

// Warn logs at warn level using the default logger.
func Warn(ctx context.Context, msg string, args ...any) {
	slog.Default().LogAttrs(ctx, slog.LevelWarn, msg, argsToAttrs(args)...)
}

// Error logs at error level using the default logger.
func Error(ctx context.Context, msg string, args ...any) {
	slog.Default().LogAttrs(ctx, slog.LevelError, msg, argsToAttrs(args)...)
}

// Debug logs at debug level using the default logger.
func Debug(ctx context.Context, msg string, args ...any) {
	slog.Default().LogAttrs(ctx, slog.LevelDebug, msg, argsToAttrs(args)...)
}

// LogAttrError returns an slog.Attr with key "error" for the given error value.
func LogAttrError(err error) slog.Attr {
	return slog.Any("error", err)
}

// LogAttrTag returns an slog.Attr with key "tag" for the given string value.
func LogAttrTag(tag string) slog.Attr {
	return slog.String("tag", tag)
}
