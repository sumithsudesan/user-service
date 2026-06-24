# notification-service

Event-driven consumer that listens for user domain events from RabbitMQ and triggers downstream actions. Currently logs each event as structured JSON. Email / SMS delivery is planned for a later phase.

## Responsibilities

- Consume user events from RabbitMQ (`user.created`, `user.updated`, `user.deleted`)
- Parse and validate the event payload
- Log each event as structured JSON via `pkg/logger`
- Acknowledge processed messages; negative-acknowledge unparseable ones

## Architecture

```text
RabbitMQ Exchange (user.events — topic, durable)
          │
          │  routing key: user.*
          ▼
Queue (notification.user.events — durable)
          │
          ▼
  pkg/queue.Consumer.Consume()
          │
          ▼
     HandleEvent()
          │
     ┌────┴─────┐
     │  success  │  → msg.Ack()   → structured JSON log
     │  failure  │  → msg.Nack()  → dead-letter (planned)
     └───────────┘
```

The service depends on `pkg/queue.Consumer` for all broker communication. No raw AMQP code lives in this service — swapping RabbitMQ for Kafka requires no changes here.

## Project Structure

```text
notification-service/
├── cmd/
│   └── main.go                  Entry point — wires consumer, handles graceful shutdown
├── src/
│   └── notification/
│       └── handler.go           Event parsing and structured JSON logging
├── tests/
│   └── handler_test.go          Unit tests for event handler
├── docker/
│   └── Dockerfile               Multi-stage Docker build (scratch final image)
├── config.yaml                  Service configuration
├── go.mod                       Module definition
└── Makefile                     Build and run targets
```

## Event Payload

```json
{
  "user_id": "user-20260624120000.000000000",
  "email": "alice@example.com",
  "event_type": "user.created",
  "timestamp": "2026-06-24T12:00:00Z"
}
```

| Event type | Trigger |
|---|---|
| `user.created` | New user registered |
| `user.updated` | User profile updated |
| `user.deleted` | User account deleted |

## Configuration

Loaded from `config.yaml`. Sensitive values override via environment variables.

```yaml
service:
  name: notification-service
  port: 0
  env: development

log:
  level: info

queue:
  provider: rabbitmq
  host: localhost
  port: 5672
  user: guest
  password: ""       # override: QUEUE_PASSWORD
  exchange:
    name: user.events
    type: topic
    durable: true
  queue:
    name: notification.user.events
    routing_key: "user.*"
    durable: true
```

## Running Locally

Requires RabbitMQ running on `localhost:5672`.

```bash
cd services/notification-service
make run
# or
go run ./cmd
```

## Running Tests

```bash
make test
# or
go test ./...
```

Tests cover:

| Test | What it verifies |
|---|---|
| `TestHandleEvent_ValidPayload` | Full valid payload is parsed and logged |
| `TestHandleEvent_InvalidJSON` | Malformed JSON returns an error |
| `TestHandleEvent_EmptyBody` | Empty JSON object parses without error |
| `TestHandleEvent_AllEventTypes` | All three event types are handled correctly |

## Docker

Build context is `services/` to include the shared `pkg` module via the Go workspace:

```bash
make docker-build        # builds image: notification-service:latest
make docker-run          # runs the container
```

Or manually from `services/`:

```bash
docker build -t notification-service:latest -f notification-service/docker/Dockerfile .
```

## Makefile Targets

| Target | Description |
|---|---|
| `make build` | Compile the service |
| `make test` | Run all tests |
| `make run` | Run locally |
| `make docker-build` | Build Docker image |
| `make docker-run` | Run Docker container |
| `make clean` | Clean build artefacts |

## Publishing a Test Event

Once running, open the RabbitMQ management UI at `http://localhost:15672` (guest / guest):

1. Go to **Exchanges** → `user.events`
2. Click **Publish message**
3. Set **Routing key**: `user.created`
4. Set **Payload**:

```json
{
  "user_id": "user-001",
  "email": "test@example.com",
  "event_type": "user.created",
  "timestamp": "2026-06-24T10:00:00Z"
}
```

Expected log output:

```json
{
  "time": "2026-06-24T10:00:01Z",
  "level": "INFO",
  "msg": "received user event",
  "service": "notification-service",
  "event_type": "user.created",
  "user_id": "user-001",
  "email": "test@example.com"
}
```

## Dependencies

| Package | Purpose |
|---|---|
| `pkg/config` | YAML + env var configuration |
| `pkg/logger` | Structured JSON logger (zap) |
| `pkg/queue` | RabbitMQ consumer abstraction |

## Current Limitations

| Limitation | Planned replacement |
|---|---|
| Logs only, no real notifications | Email / SMS / webhook delivery |
| No retry on failed messages | Dead-letter queue + retry policy |
| No reconnect on broker failure | Reconnect with exponential backoff |
