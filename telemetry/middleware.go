package telemetry

import (
	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// GinRequestIDMiddleware injects a request ID into the gin context.
// Reads from the header key (default: x-request-id), generates a UUID v7 if missing.
func GinRequestIDMiddleware(key ...string) gin.HandlerFunc {
	k := resolveKey(key...)
	return func(c *gin.Context) {
		requestID := c.GetHeader(k)
		if requestID == "" {
			requestID = generateRequestID()
		}

		ctx := SetRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Header(k, requestID)

		c.Next()
	}
}

// EchoRequestIDMiddleware injects a request ID into the echo context.
// Reads from the header key (default: x-request-id), generates a UUID v7 if missing.
func EchoRequestIDMiddleware(key ...string) echo.MiddlewareFunc {
	k := resolveKey(key...)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Request().Header.Get(k)
			if requestID == "" {
				requestID = generateRequestID()
			}

			ctx := SetRequestID(c.Request().Context(), requestID)
			c.SetRequest(c.Request().WithContext(ctx))
			c.Response().Header().Set(k, requestID)

			return next(c)
		}
	}
}

// FiberRequestIDMiddleware injects a request ID into the fiber context.
// Reads from the header key (default: x-request-id), generates a UUID v7 if missing.
func FiberRequestIDMiddleware(key ...string) fiber.Handler {
	k := resolveKey(key...)
	return func(c *fiber.Ctx) error {
		requestID := c.Get(k)
		if requestID == "" {
			requestID = generateRequestID()
		}

		ctx := SetRequestID(c.UserContext(), requestID)
		c.SetUserContext(ctx)
		c.Set(k, requestID)

		return c.Next()
	}
}

func resolveKey(key ...string) string {
	if len(key) > 0 && key[0] != "" {
		return key[0]
	}
	return XRequestIdKey
}

func GinMiddleware(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}

func GinMiddlewareWithConfig(serviceName string, opts ...otelgin.Option) gin.HandlerFunc {
	return otelgin.Middleware(serviceName, opts...)
}

func CustomGinMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		propagator := propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)

		ctx = propagator.Extract(ctx, propagation.HeaderCarrier(c.Request.Header))

		ctx, span := StartSpan(ctx, c.Request.Method+" "+c.FullPath(),
			oteltrace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.url", c.Request.URL.String()),
				attribute.String("http.scheme", c.Request.URL.Scheme),
				attribute.String("http.host", c.Request.Host),
				attribute.String("http.target", c.Request.URL.Path),
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("http.client_ip", c.ClientIP()),
			),
		)
		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		traceID := TraceID(ctx)
		spanID := SpanID(ctx)

		c.Header("X-Trace-ID", traceID)
		c.Header("X-Span-ID", spanID)

		propagator.Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))

		c.Next()

		status := c.Writer.Status()
		span.SetAttributes(
			attribute.Int("http.status_code", status),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		if status >= 400 {
			span.SetStatus(codes.Error, "HTTP Error")
		} else {
			span.SetStatus(codes.Ok, "")
		}

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				RecordError(ctx, err.Err)
			}
		}
	}
}
