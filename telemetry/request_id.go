package telemetry

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const XRequestIdKey = "x-request-id"

type contextKey string

const requestIDContextKey contextKey = XRequestIdKey

type Tracer interface {
	AutoTracingMiddleware() func(http.Handler) http.Handler
}

type tracer struct {
	requestIdKey string
}

type TracerConfig struct {
	RequestIdKey string
}

func NewTracer(cfg TracerConfig) Tracer {
	if cfg.RequestIdKey == "" {
		cfg.RequestIdKey = XRequestIdKey
	}

	return &tracer{
		requestIdKey: cfg.RequestIdKey,
	}
}

func (t *tracer) AutoTracingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(t.requestIdKey)
			if requestID == "" {
				requestID = generateRequestID()
			}

			ctx := SetRequestID(r.Context(), requestID)
			w.Header().Set(t.requestIdKey, requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func GetRequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDContextKey).(string)
	return id
}

func generateRequestID() string {
	id, err := uuid.NewV7()
	if err != nil {
		id = uuid.New()
	}
	return id.String()
}
