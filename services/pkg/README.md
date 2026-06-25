# pkg

Shared library module for the Cloud-Native User Platform. Provides reusable infrastructure abstractions consumed by all services via a Go workspace.

All packages expose interfaces — services depend on the interface, not the implementation. Adding a new provider (e.g. Kafka, MySQL) requires only a new implementation file with no changes to business logic.

## Packages

### config

Loads configuration from a YAML file with environment variable overrides (via Viper).

```go
cfg, err := config.Load("config.yaml")
// cfg.Service.Name, cfg.Log.Level, cfg.Database, cfg.Queue ...
```

Environment variables use `_` as a separator and are automatically mapped:
- `DATABASE_PASSWORD` → `database.password`
- `QUEUE_PASSWORD` → `queue.password`

### logger

Structured JSON logger interface backed by zap.

```go
log, err := logger.New("info")  // levels: debug | info | warn | error
log.Info("user created", "user_id", id)
log.Error("db error", "error", err)
```

**Interface:**

```go
type Logger interface {
    Debug(msg string, keysAndValues ...any)
    Info(msg string, keysAndValues ...any)
    Warn(msg string, keysAndValues ...any)
    Error(msg string, keysAndValues ...any)
    With(keysAndValues ...any) Logger
}
```

### database

PostgreSQL connection abstraction using pgx.

```go
db, err := database.New(ctx, cfg.Database, log)
defer db.Close()
// db.Pool() returns *pgxpool.Pool for query execution
```

Configured via `DatabaseConfig` (host, port, name, user, password, pool settings, timeouts).

### queue

Message queue abstraction with Publisher and Consumer interfaces.

```go
// Publishing (user-service)
pub, err := queue.NewPublisher(cfg.Queue, log)
pub.Publish(ctx, queue.Message{RoutingKey: "user.created", Body: payload})

// Consuming (notification-service)
con, err := queue.NewConsumer(cfg.Queue, log)
con.Consume(ctx, func(msg queue.Message) error {
    // handle msg.Body
    return nil
})
```

**Interfaces:**

```go
type Publisher interface {
    Publish(ctx context.Context, msg Message) error
    Close() error
}

type Consumer interface {
    Consume(ctx context.Context, handler HandlerFunc) error
    Close() error
}

type HandlerFunc func(msg Message) error
```

Current provider: **RabbitMQ**. Add new providers by implementing the interface and adding a case to `NewPublisher` / `NewConsumer`.

## Configuration Model

All packages are configured through the shared `Config` struct:

```text
Config
├── Service   name, port, env
├── Log       level
├── Database  provider, host, port, name, user, password, pool, timeout
└── Queue     provider, host, port, user, password, exchange, queue, retry, dlq
```

## Usage in Services

Services import `pkg` packages directly. The Go workspace (`services/go.work`) resolves the local module without requiring a published version:

```go
import (
    "github.com/sumithsudesan/pkg/config"
    "github.com/sumithsudesan/pkg/logger"
    "github.com/sumithsudesan/pkg/database"
    "github.com/sumithsudesan/pkg/queue"
)
```

## Running Tests

```bash
cd services/pkg
go test ./...
```
