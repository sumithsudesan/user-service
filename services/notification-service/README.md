# Notification Service

The notification service is an event-driven consumer that listens for user domain events
published by the user service via RabbitMQ. Its responsibility is to react to those events
and trigger downstream actions — currently structured JSON logging, with email/SMS delivery
planned in a later phase.

## Responsibilities

- Consume user events from RabbitMQ (`user.created`, `user.updated`, `user.deleted`)
- Parse and validate the event payload
- Log each event as structured JSON
- Acknowledge processed messages; negative-acknowledge unparseable ones

---

## Architecture

```text
RabbitMQ Exchange (user.events — topic, durable)
          │
          │  routing key: user.*
          ▼
Queue (notification.user.events — durable)
          │
          ▼
    Consumer.Start()
          │
          ▼
    HandleEvent()
          │
     ┌────┴─────┐
     │  success  │  → msg.Ack()  → structured JSON log
     │  failure  │  → msg.Nack() → dead-letter (future)
     └───────────┘
```

---

## Event Payload

Events published by the user service follow this format:

```json
{
  "user_id": "user-20260624120000.000000000",
  "email": "user@example.com",
  "event_type": "user.created",
  "timestamp": "2026-06-24T10:00:00Z"
}
```

Supported event types:

| Event type       | Trigger                    |
|------------------|----------------------------|
| `user.created`   | New user registered        |
| `user.updated`   | User profile updated       |
| `user.deleted`   | User account deleted       |

---

## Configuration

| Environment Variable | Default                                | Description             |
|----------------------|----------------------------------------|-------------------------|
| `AMQP_URL`           | `amqp://guest:guest@localhost:5672/`   | RabbitMQ connection URL |

---

## Project Structure

```text
notification-service/
├── cmd/
│   └── main.go                  Entry point — wires consumer, handles graceful shutdown
├── src/
│   └── notification/
│       ├── consumer.go          RabbitMQ connection, exchange/queue setup, consume loop
│       └── handler.go           Event parsing and structured JSON logging
├── tests/
│   └── handler_test.go          Unit tests for event handler
└── docker/
    └── Dockerfile               Multi-stage Docker build
```

---

## Running Locally

Requires RabbitMQ running on `localhost:5672`.

```bash
cd services/notification-service
go run ./cmd/main.go
```

Override the RabbitMQ URL:

```bash
AMQP_URL=amqp://user:pass@localhost:5672/ go run ./cmd/main.go
```

Or via the Makefile from the repo root:

```bash
make run-notification
```

---

## Running Tests

```bash
make test-notification
```

Or directly:

```bash
cd services/notification-service
go test -vet=off ./tests/...
```

Tests cover:

| Test                              | What it verifies                            |
|-----------------------------------|---------------------------------------------|
| `TestHandleEvent_ValidPayload`    | Full valid payload is parsed and logged     |
| `TestHandleEvent_InvalidJSON`     | Malformed JSON returns an error             |
| `TestHandleEvent_EmptyBody`       | Empty JSON object parses without error      |
| `TestHandleEvent_AllEventTypes`   | All three event types are handled correctly |

---

## Docker

Build the image:

```bash
make build-notification
```

Build with a specific tag and registry (for CI):

```bash
make build-notification TAG=abc1234 REGISTRY=ghcr.io/sumithsudesan
```

Start all services via docker compose:

```bash
make up          # starts RabbitMQ + notification-service
make logs        # tail notification-service output
make down        # stop everything
```

---

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

Expected log output from the notification service:

```json
{
  "time": "2026-06-24T10:00:01Z",
  "level": "INFO",
  "msg": "received user event",
  "service": "notification-service",
  "event_type": "user.created",
  "user_id": "user-001",
  "email": "test@example.com",
  "timestamp": "2026-06-24T10:00:00Z"
}
```

---

## Dependencies

| Package                      | Purpose               |
|------------------------------|-----------------------|
| `rabbitmq/amqp091-go v1.10` | RabbitMQ AMQP client  |

---

## Current Limitations

This is the Phase 1 implementation. The following are planned for later phases:

| Limitation                            | Planned replacement                        |
|---------------------------------------|--------------------------------------------|
| AMQP URL read from env var only       | `pkg/config` — YAML + env var override     |
| `log/slog` stdlib logger              | `pkg/logger` — zap structured JSON logger  |
| Logs only, no real notifications      | Email / SMS / webhook delivery             |
| No retry on failed messages           | Dead-letter queue + retry policy           |
| No connection reconnect on failure    | Reconnect with exponential backoff         |
