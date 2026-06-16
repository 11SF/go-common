# go-common

Shared Go library for internal services. Module path: `github.com/11SF/go-common`.

## Package overview

| Package | Import path | Purpose |
|---|---|---|
| `config` | `github.com/11SF/go-common/config` | Load env vars into a typed struct |
| `logger` | `github.com/11SF/go-common/logger` | Structured JSON logging (slog-based) |
| `kafka` | `github.com/11SF/go-common/kafka` | Kafka producer + consumer group |
| `redis` | `github.com/11SF/go-common/redis` | Redis client + Streams message queue |
| `database` | `github.com/11SF/go-common/database` | Generic DB connection (Gorm / Bun / sql.DB) |
| `postgres` | `github.com/11SF/go-common/postgres` | Postgres-specific helpers |
| `s3` | `github.com/11SF/go-common/s3` | S3-compatible object storage (AWS, MinIO, DO) |
| `http_client` | `github.com/11SF/go-common/http_client` | Typed HTTP client (fasthttp-based) |
| `shutdown` | `github.com/11SF/go-common/shutdown` | Graceful shutdown for Gin / Echo / Fiber |
| `jwt` | `github.com/11SF/go-common/jwt` | JWT sign / verify |
| `response` | `github.com/11SF/go-common/response` | Standard HTTP response envelope |
| `context` | `github.com/11SF/go-common/context` | User context helpers |
| `telemetry` | `github.com/11SF/go-common/telemetry` | OpenTelemetry tracer setup |

---

## config

Reads environment variables into any struct using `env:` tags. Loads `.env` file first (errors are non-fatal).

```go
import commonconfig "github.com/11SF/go-common/config"

type AppConfig struct {
    Port     string `env:"PORT" envDefault:"8080"`
    DBHost   string `env:"DB_HOST"`
}

cfg := commonconfig.LoadConfig[AppConfig](ctx, ".env")
// cfg is *AppConfig; panics if required fields are missing
```

---

## logger

Structured JSON logger backed by `log/slog`. Automatically injects `trace_id` / `span_id` from context.

```go
import "github.com/11SF/go-common/logger"

// Init once at startup
logger.Init("INFO") // INFO | DEBUG | WARN | ERROR

// Use anywhere
logger.Info(ctx, "user created", slog.String("user_id", id))
logger.Error(ctx, "db error", logger.LogAttrError(err), logger.LogAttrTag("user-service"))
logger.Debug(ctx, "payload", slog.Any("body", body))
```

Helper constructors:
- `logger.LogAttrError(err)` → `slog.Attr{Key:"err"}`
- `logger.LogAttrTag(tag)` → `slog.Attr{Key:"tag"}`

---

## kafka

Producer + consumer group with SASL (PLAIN / SCRAM-SHA-256 / SCRAM-SHA-512), TLS, Snappy compression, in-memory retry, and dead-letter topic.

### Shared Config

```go
import "github.com/11SF/go-common/kafka"

cfg := kafka.Config{
    Brokers:       []string{"broker1:9092", "broker2:9092"},
    SASLEnable:    true,
    SASLMechanism: kafka.SASLSCRAMSHA512, // kafka.SASLPlain | kafka.SASLSCRAMSHA256 | kafka.SASLSCRAMSHA512
    SASLUser:      "user",
    SASLPassword:  "pass",
    TLSEnable:     true,
}
```

Env vars (used with `config.LoadConfig`):
- `KAFKA_BROKERS` (comma-separated)
- `KAFKA_SASL_ENABLE`, `KAFKA_SASL_MECHANISM`, `KAFKA_SASL_USER`, `KAFKA_SASL_PASSWORD`
- `KAFKA_TLS_ENABLE`

### Producer

```go
producer, err := kafka.NewProducer(ctx, kafka.ProducerConfig{
    Config:       cfg,
    Idempotent:   true,    // exactly-once; requires RequiredAcks=WaitForAll (default)
    Compression:  sarama.CompressionSnappy, // default; omit to keep default
    MaxRetry:     3,       // default
    RetryBackoff: 250 * time.Millisecond, // default
})
if err != nil { ... }
defer producer.Close()

// Send raw bytes
partition, offset, err := producer.Send(ctx, "orders", []byte("key"), []byte("value"))

// Send JSON
err = producer.SendJSON(ctx, "orders", "order-123", myStruct)
```

### Consumer

```go
consumer, err := kafka.NewConsumer(ctx, kafka.ConsumerConfig{
    Config:          cfg,
    Group:           "my-service",
    Topics:          []string{"orders"},
    OffsetOldest:    true,            // false = OffsetNewest (default)
    MaxRetry:        3,               // in-memory retries before dead-letter
    RetryBackoff:    500 * time.Millisecond,
    DeadLetterTopic: "orders.dead-letter", // omit to skip DLQ
})
if err != nil { ... }
defer consumer.Close()

// Subscribe blocks until ctx is cancelled
err = consumer.Subscribe(ctx, func(ctx context.Context, msg kafka.Message) error {
    // msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value, msg.Headers, msg.Timestamp
    var order Order
    if err := json.Unmarshal(msg.Value, &order); err != nil {
        return err // will be retried
    }
    return processOrder(ctx, order)
})
```

Consumer behaviour:
- Retries the handler up to `MaxRetry` times with `RetryBackoff` sleep between attempts.
- After exhausting retries, forwards to `DeadLetterTopic` (if configured) with `_source_topic`, `_source_offset`, `_error` headers.
- Always commits the offset after processing so a bad message never blocks the partition.
- Handles consumer group rebalances automatically.

---

## redis

### Client

```go
import redisclient "github.com/11SF/go-common/redis"

client := redisclient.NewRedisClient(ctx, redisclient.Config{
    Addr:     "localhost:6379",
    Password: "",
})
// returns redis.UniversalClient; panics on connection failure
```

Env vars: `REDIS_ADDRESS`, `REDIS_PASSWORD`

### Message Queue (Redis Streams)

```go
mq, err := redisclient.NewMessageQueue(ctx, client, redisclient.StreamConfig{
    Stream:   "events",
    Group:    "my-service",
    Consumer: "instance-1",
    // BatchSize: 10, BlockDuration: 5s, MaxRetry: 3 (defaults)
})

// Publish
id, err := mq.Publish(ctx, map[string]any{"event": "order.created", "id": "123"})

// Subscribe (blocking)
err = mq.Subscribe(ctx, func(ctx context.Context, msg redisclient.Message) error {
    fmt.Println(msg.ID, msg.Values)
    return nil
})
```

Failed messages (handler returns error) are retried up to `MaxRetry` times, then moved to `<stream>:dead-letter`.

---

## database

Low-level DB connection helpers. Prefer using `postgres` package for Postgres-specific setup.

```go
import "github.com/11SF/go-common/database"

// sql.DB (used as base for Bun)
db := database.NewDatabase(ctx, database.Config{
    Host: "localhost", Port: 5432,
    Username: "user", Password: "pass", DatabaseName: "mydb",
    // SSLMode, MaxOpenConns, MaxIdleConns are *int/*string (optional)
})

// Gorm
dial, _ := postgres.ConnectPostgres(&postgres.Config{...})
gormDB, err := database.InitGormDatabase(&database.GormConfig{Dial: dial})

// Bun
bunDB := database.NewBunDatabase(&database.BunConfig{
    Sql: db, Dialect: pgdialect.New(),
})
```

Env vars: `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD`, `DB_NAME`, `DB_SSL_MODE`, `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`

---

## s3

Supports AWS S3, MinIO, DigitalOcean Spaces, and any S3-compatible store. Auto-detects path-style, region defaults, and SSL from the endpoint URL.

```go
import "github.com/11SF/go-common/s3"

client, err := s3.NewClient(&s3.Config{
    Provider:        s3.ProviderAWS,   // ProviderAWS | ProviderMinIO | ProviderDigitalOcean | ProviderCustom
    Region:          "ap-southeast-1",
    AccessKeyID:     "key",
    SecretAccessKey: "secret",
    BucketName:      "my-bucket",
    // Endpoint: "" for AWS; set for MinIO e.g. "http://localhost:9000"
})

// Operations available on *s3.Client (see s3/operations.go)
```

---

## http_client

Generic typed HTTP client backed by `fasthttp`.

```go
import httpclient "github.com/11SF/go-common/http_client"

client := httpclient.Client{
    Client:        &fasthttp.Client{},
    Timeout:       5 * time.Second,
    EnableLogging: true,
    DefaultHeaderOptions: []httpclient.HeaderOption{
        httpclient.WithBearerToken("token"),
    },
}

// GET
res, err := httpclient.Get[MyResponse](client, "https://api.example.com/users", map[string]string{"page": "1"})
// res.Code, res.Response (typed)

// POST
res, err := httpclient.Post[MyRequest, MyResponse](client, url, myRequestBody)

// DELETE
res, err := httpclient.Delete[MyResponse](client, url)
```

---

## shutdown

Graceful shutdown for HTTP servers. Listens for `SIGTERM` / `SIGINT`.

```go
import "github.com/11SF/go-common/shutdown"

cfg := shutdown.Config{Addr: ":8080", Timeout: 10 * time.Second}

// Pick the matching helper for your framework:
shutdown.GracefulShutdownGin(ctx, cfg, ginRouter)
shutdown.GracefulShutdownEcho(ctx, cfg, echoInstance)
shutdown.GracefulShutdownFiber(ctx, cfg, fiberApp)
```

---

## Patterns used across all packages

- **Config structs carry `env:` tags** — pass them to `config.LoadConfig` and the fields are filled from env vars automatically.
- **`NewXxx(ctx, cfg)` panics or returns error** — connection failures that should crash the service (Redis, DB) call `panic`; others return `(T, error)`.
- **Logger is always the package-level `logger.Info/Error/Warn/Debug`** — never a struct field. Pass `ctx` so trace IDs propagate.
- **Graceful shutdown** — all blocking loops (`Subscribe`, `Subscribe`) respect `ctx.Done()` and exit cleanly.
