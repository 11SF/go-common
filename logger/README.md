# Logger Package

Structured JSON logging with OpenTelemetry tracing integration for observability and debugging.

## Features

- ✅ JSON structured logging with customizable fields
- ✅ OpenTelemetry trace/span ID integration
- ✅ Context-aware logging with automatic trace correlation
- ✅ Multiple log levels (Info, Warn, Error, Debug)
- ✅ Gin middleware for HTTP request logging
- ✅ Fallback to custom tracing if OpenTelemetry unavailable

## Quick Start

### 1. Initialize Logger

```go
import "your-project/pkg/logger"

func init() {
    // Initialize logger (sets as default slog logger)
    logger.Init()
}
```

### 2. Add Gin Middleware

```go
import (
    "your-project/pkg/logger"
    "your-project/pkg/telemetry"
)

func setupRouter() *gin.Engine {
    router := gin.New()

    // Add telemetry middleware first (for trace context)
    router.Use(telemetry.GinMiddleware("your-service"))

    // Add logger middleware (will use telemetry trace IDs)
    router.Use(logger.GinMiddleware())

    return router
}
```

### 3. Context-Aware Logging

```go
func yourHandler(c *gin.Context) {
    ctx := c.Request.Context()

    // Logs will automatically include trace_id and span_id
    logger.Info(ctx, "Processing user request",
        "user_id", 123,
        "action", "create_order")

    if err := processOrder(ctx); err != nil {
        logger.Error(ctx, "Failed to process order",
            "error", err.Error(),
            "user_id", 123)
        return
    }

    logger.Info(ctx, "Order processed successfully")
}
```

## API Reference

### Global Functions

```go
// Context-aware logging (recommended)
func Info(ctx context.Context, msg string, args ...any)
func Warn(ctx context.Context, msg string, args ...any)
func Error(ctx context.Context, msg string, args ...any)
func Debug(ctx context.Context, msg string, args ...any)

// Initialize logger
func Init()

// Create new logger instance
func New() *Logger
```

### Logger Instance Methods

```go
// Create logger with tracing context
func (l *Logger) WithTracing(ctx context.Context) *slog.Logger

// Get instance for Gin
func GinLogger() *Logger
```

### Middleware

```go
// Gin middleware for HTTP request logging
func GinMiddleware() gin.HandlerFunc
```

### Custom Tracing (Fallback)

```go
// Manual trace context creation (when OpenTelemetry unavailable)
func NewTraceContext(ctx context.Context) context.Context
func WithTraceID(ctx context.Context, traceID string) context.Context
func WithSpanID(ctx context.Context, spanID string) context.Context

// Get trace IDs
func GetTraceID(ctx context.Context) string
func GetSpanID(ctx context.Context) string

// Generate IDs
func GenerateTraceID() string // 32 hex chars
func GenerateSpanID() string  // 16 hex chars
```

## JSON Output Format

### Standard Log Entry

```json
{
  "timestamp": "2024-01-01T12:00:00.123Z",
  "level": "INFO",
  "message": "Processing user request",
  "trace_id": "ba68b049e7aa5b6aa843072aebab45ee",
  "span_id": "bdc0fdca899a95e3",
  "user_id": 123,
  "action": "create_order"
}
```

### HTTP Request Log

```json
{
  "timestamp": "2024-01-01T12:00:00.456Z",
  "level": "INFO",
  "message": "HTTP Request",
  "trace_id": "ba68b049e7aa5b6aa843072aebab45ee",
  "span_id": "bdc0fdca899a95e3",
  "method": "GET",
  "path": "/api/users/123",
  "status": 200,
  "latency": "45.2ms",
  "client_ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "body_size": 1024
}
```

### Error Log

```json
{
  "timestamp": "2024-01-01T12:00:00.789Z",
  "level": "ERROR",
  "message": "Database connection failed",
  "trace_id": "ba68b049e7aa5b6aa843072aebab45ee",
  "span_id": "bdc0fdca899a95e3",
  "error": "connection timeout",
  "database": "users_db",
  "retry_count": 3
}
```

## Usage Examples

### Basic Logging

```go
func processUser(ctx context.Context, userID int) error {
    logger.Info(ctx, "Starting user processing", "user_id", userID)

    user, err := database.GetUser(ctx, userID)
    if err != nil {
        logger.Error(ctx, "Failed to fetch user",
            "user_id", userID,
            "error", err)
        return err
    }

    logger.Info(ctx, "User fetched successfully",
        "user_id", userID,
        "username", user.Name)

    return nil
}
```

### HTTP Handler with Tracing

```go
func getUserHandler(c *gin.Context) {
    ctx := c.Request.Context()
    userID := c.Param("id")

    logger.Info(ctx, "Get user request", "user_id", userID)

    // Create child span for database operation
    ctx, span := telemetry.StartSpan(ctx, "database-get-user")
    defer span.End()

    user, err := userService.GetUser(ctx, userID)
    if err != nil {
        logger.Error(ctx, "Failed to get user",
            "user_id", userID,
            "error", err)

        c.JSON(500, gin.H{"error": "Internal server error"})
        return
    }

    logger.Info(ctx, "User retrieved successfully",
        "user_id", userID)

    c.JSON(200, user)
}
```

### Background Job Logging

```go
func processEmailQueue(ctx context.Context) {
    // Create new trace context for background job
    ctx = logger.NewTraceContext(ctx)

    logger.Info(ctx, "Starting email queue processing")

    emails, err := emailQueue.GetPending(ctx)
    if err != nil {
        logger.Error(ctx, "Failed to fetch pending emails", "error", err)
        return
    }

    for _, email := range emails {
        // Create child context for each email
        emailCtx := logger.WithSpanID(ctx, logger.GenerateSpanID())

        logger.Info(emailCtx, "Processing email",
            "email_id", email.ID,
            "recipient", email.To)

        if err := sendEmail(emailCtx, email); err != nil {
            logger.Error(emailCtx, "Failed to send email",
                "email_id", email.ID,
                "error", err)
            continue
        }

        logger.Info(emailCtx, "Email sent successfully",
            "email_id", email.ID)
    }

    logger.Info(ctx, "Email queue processing completed",
        "processed_count", len(emails))
}
```

## Configuration

### Field Customization

The logger uses custom field names for better readability:

```go
// Default slog fields → Custom fields
"time" → "timestamp"
"level" → "level"
"msg" → "message"
```

### Log Levels

```go
// Set log level (default: Info)
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug, // Debug, Info, Warn, Error
}))
```

## Integration Patterns

### With OpenTelemetry

```go
// Automatic integration - no additional setup needed
router.Use(telemetry.GinMiddleware("service-name"))
router.Use(logger.GinMiddleware())

// Logs will automatically include OpenTelemetry trace/span IDs
```

### Without OpenTelemetry (Fallback)

```go
// Manual trace context creation
func processRequest(ctx context.Context) {
    ctx = logger.NewTraceContext(ctx)

    logger.Info(ctx, "Processing request")
    // Will generate custom trace/span IDs
}
```

### With Multiple Services

```go
// Service A
func callServiceB(ctx context.Context) {
    logger.Info(ctx, "Calling service B")

    // HTTP client with trace context
    req, _ := http.NewRequestWithContext(ctx, "GET", "http://service-b/api", nil)

    // Service B will continue the same trace
    resp, err := client.Do(req)

    logger.Info(ctx, "Service B response received", "status", resp.StatusCode)
}
```

## Best Practices

### 1. Structured Logging

```go
// ✅ Good: Use key-value pairs
logger.Info(ctx, "User created",
    "user_id", user.ID,
    "email", user.Email,
    "role", user.Role)

// ❌ Avoid: String interpolation
logger.Info(ctx, fmt.Sprintf("User %d created with email %s", user.ID, user.Email))
```

### 2. Error Context

```go
// ✅ Good: Include relevant context
logger.Error(ctx, "Payment processing failed",
    "user_id", payment.UserID,
    "amount", payment.Amount,
    "payment_id", payment.ID,
    "error", err.Error())

// ❌ Avoid: Minimal context
logger.Error(ctx, "Error occurred", "error", err.Error())
```

### 3. Performance Considerations

```go
// ✅ Good: Log at appropriate levels
logger.Debug(ctx, "Detailed debugging info") // Only in debug mode
logger.Info(ctx, "Important business events") // Always logged
logger.Error(ctx, "Critical errors") // Always logged

// ❌ Avoid: Excessive logging in hot paths
for _, item := range millionItems {
    logger.Info(ctx, "Processing item", "item_id", item.ID) // Too verbose
}
```

### 4. Sensitive Data

```go
// ✅ Good: Mask sensitive information
logger.Info(ctx, "Login attempt",
    "username", username,
    "ip", clientIP,
    "success", true)

// ❌ Avoid: Logging passwords, tokens, etc.
logger.Info(ctx, "Login successful",
    "username", username,
    "password", password) // Never log passwords!
```

## Troubleshooting

### Missing Trace IDs

If trace IDs are not appearing in logs:

1. Ensure telemetry middleware is added before logger middleware
2. Check that OpenTelemetry is properly initialized
3. Verify context is being passed correctly

### Log Format Issues

For ELK/Loki integration, ensure JSON format:

```bash
# Verify JSON output
go run cmd/main.go 2>&1 | jq '.'
```

### Performance Impact

Monitor logging performance in high-traffic scenarios:

```go
// Consider async logging for high volume
// Use appropriate log levels
// Avoid logging in tight loops
```