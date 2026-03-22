package httpclient

import (
	"context"

	"github.com/11SF/go-common/telemetry"
)

type HeaderOption func() (key, value string)

const (
	ContentTypeJSON           = "application/json"
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
)

func WithHeader(key, value string) HeaderOption {
	return func() (string, string) {
		return key, value
	}
}

// --- Common helpers ---

func WithBearerToken(token string) HeaderOption {
	return func() (string, string) {
		return "Authorization", "Bearer " + token
	}
}

func WithContentType(ct string) HeaderOption {
	return func() (string, string) {
		return "Content-Type", ct
	}
}

func WithContentTypeJSON() HeaderOption {
	return func() (string, string) {
		return "Content-Type", ContentTypeJSON
	}
}

func WithRequestID(id string) HeaderOption {
	return func() (string, string) {
		return "X-Request-ID", id
	}
}

func WithAcceptJSON() HeaderOption {
	return func() (string, string) {
		return "Accept", "application/json"
	}
}

func WithTracing(ctx context.Context) HeaderOption {
	return func() (string, string) {
		reqId := telemetry.GetRequestID(ctx)
		return telemetry.XRequestIdKey, reqId
	}
}

func WithTracingCustomKey(ctx context.Context, tracingKey string) HeaderOption {
	return func() (string, string) {
		reqId := telemetry.GetRequestID(ctx)
		return tracingKey, reqId
	}
}
