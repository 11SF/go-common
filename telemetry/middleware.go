package telemetry

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

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
