# Telemetry Package

OpenTelemetry tracing implementation for distributed tracing and observability.

## Features

- ✅ W3C Trace Context compliant (32-char trace ID, 16-char span ID)
- ✅ Multiple exporters support (stdout, Jaeger, Zipkin)
- ✅ Automatic HTTP instrumentation with Gin middleware
- ✅ Context propagation across services
- ✅ Standard semantic attributes
- ✅ Error recording and span status tracking

## Quick Start

### 1. Initialize Telemetry

```go
import "your-project/pkg/telemetry"

func init() {
    err := telemetry.Init(telemetry.Config{
        ServiceName:    "your-service",
        ServiceVersion: "1.0.0",
        Environment:    "development",
        ExporterType:   "stdout", // or "jaeger"
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

### 2. Add Gin Middleware

```go
import "your-project/pkg/telemetry"

func setupRouter() *gin.Engine {
    router := gin.New()

    // Add OpenTelemetry middleware
    router.Use(telemetry.GinMiddleware("your-service"))

    return router
}
```

### 3. Create Custom Spans

```go
func businessLogic(ctx context.Context) error {
    // Create a child span
    ctx, span := telemetry.StartSpan(ctx, "database-operation")
    defer span.End()

    // Your business logic here
    result, err := database.Query(ctx, "SELECT * FROM users")

    if err != nil {
        // Record error in span
        telemetry.RecordError(ctx, err)
        return err
    }

    // Add custom event
    telemetry.AddEvent(ctx, "query-completed",
        attribute.Int("result_count", len(result)))

    return nil
}
```

## Configuration

### Environment Variables

```bash
# Service configuration
ENVIRONMENT=development|staging|production
OTEL_EXPORTER=stdout|jaeger|zipkin

# Jaeger specific
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
```

### Config Options

```go
type Config struct {
    ServiceName    string // Required: service identifier
    ServiceVersion string // Service version for resource attributes
    Environment    string // deployment environment
    ExporterType   string // "stdout", "jaeger", "zipkin"
}
```

## API Reference

### Global Functions

```go
// Initialize telemetry
func Init(config Config) error

// Get global instance
func Global() *Telemetry

// Start a new span
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)

// Get trace/span IDs from context
func TraceID(ctx context.Context) string
func SpanID(ctx context.Context) string

// Record error and events
func RecordError(ctx context.Context, err error, opts ...trace.EventOption)
func AddEvent(ctx context.Context, name string, opts ...trace.EventOption)
```

### Middleware Options

```go
// Standard Gin middleware
func GinMiddleware(serviceName string) gin.HandlerFunc

// Middleware with custom options
func GinMiddlewareWithConfig(serviceName string, opts ...otelgin.Option) gin.HandlerFunc

// Custom middleware with additional features
func CustomGinMiddleware(serviceName string) gin.HandlerFunc
```

## Output Examples

### Development (stdout exporter)

```json
{
  "Name": "GET /api/users",
  "SpanContext": {
    "TraceID": "ba68b049e7aa5b6aa843072aebab45ee",
    "SpanID": "bdc0fdca899a95e3"
  },
  "Attributes": [
    {"Key": "http.method", "Value": "GET"},
    {"Key": "http.route", "Value": "/api/users"},
    {"Key": "http.status_code", "Value": 200}
  ]
}
```

### Production (Jaeger)

Traces will be sent to Jaeger for:
- 🔍 Visual trace timeline
- 📊 Service dependency mapping
- 🎯 Performance bottleneck analysis
- 🚨 Error tracking and alerting

## Best Practices

### Span Naming

```go
// ✅ Good: Action + Resource
telemetry.StartSpan(ctx, "database-query-users")
telemetry.StartSpan(ctx, "http-call-auth-service")

// ❌ Avoid: Too generic
telemetry.StartSpan(ctx, "function")
telemetry.StartSpan(ctx, "processing")
```

### Error Handling

```go
ctx, span := telemetry.StartSpan(ctx, "payment-processing")
defer span.End()

if err := processPayment(ctx, amount); err != nil {
    // Record error with context
    telemetry.RecordError(ctx, err)

    // Set span status
    span.SetStatus(codes.Error, "Payment processing failed")
    return err
}

span.SetStatus(codes.Ok, "Payment completed successfully")
```

### Attributes and Events

```go
// Add meaningful attributes
span.SetAttributes(
    attribute.String("user.id", userID),
    attribute.Float64("payment.amount", amount),
    attribute.String("payment.currency", "USD"),
)

// Add events for important milestones
telemetry.AddEvent(ctx, "payment-validated")
telemetry.AddEvent(ctx, "payment-processed")
```

## Integration with Other Services

### Cross-Service Tracing

The middleware automatically handles W3C trace context propagation:

```go
// Service A
func callServiceB(ctx context.Context) {
    ctx, span := telemetry.StartSpan(ctx, "call-service-b")
    defer span.End()

    // HTTP client will automatically propagate trace context
    req, _ := http.NewRequestWithContext(ctx, "GET", "http://service-b/api", nil)
    resp, err := client.Do(req)
    // ...
}

// Service B will automatically continue the same trace
```

### Logging Integration

Use with the logger package for correlated logs:

```go
import (
    "your-project/pkg/telemetry"
    "your-project/pkg/logger"
)

func handler(c *gin.Context) {
    ctx := c.Request.Context()

    // Logs will automatically include trace_id and span_id
    logger.Info(ctx, "Processing request")

    ctx, span := telemetry.StartSpan(ctx, "business-logic")
    defer span.End()

    logger.Info(ctx, "Business logic completed")
}
```

## Troubleshooting

### Common Issues

1. **Missing traces**: Ensure telemetry.Init() is called before any span creation
2. **No trace propagation**: Check that middleware is added before other middlewares
3. **Performance impact**: Use sampling in production (configure in tracer provider)

### Debug Mode

Enable debug logging to troubleshoot:

```go
// Add debug exporter alongside main exporter
exporter := stdouttrace.New(stdouttrace.WithPrettyPrint())
```