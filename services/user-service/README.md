# user-service

REST API for user lifecycle management. Core domain service of the Cloud-Native User Platform.

## Responsibilities

- Create, read, update, and delete users via HTTP
- Enforce domain validation rules
- Publish domain events (`user.created`, `user.updated`, `user.deleted`) to RabbitMQ
- Optimistic concurrency control via version field

## Architecture

```text
HTTP Request
    │
    ▼
Router (chi)
    │
    ▼
Handler  ── request parsing, response serialisation
    │
    ▼
Service  ── domain logic, validation, business rules
    │              │
    │              └──── Publisher Interface ──── RabbitMQ (pkg/queue)
    ▼
Repository Interface
    │
    ▼
PostgreSQL (pkg/database)
```

Business logic never depends on infrastructure implementations. All external concerns (database, queue, config, logger) are injected as interfaces from `pkg`.

## Domain Model

```text
User
├── id          string     Unique identifier (user-<timestamp>)
├── name        string     Display name
├── email       string     Contact email
├── status      string     Account status (active / inactive)
├── created_at  time.Time  UTC creation timestamp
├── updated_at  time.Time  UTC last-modified timestamp
└── version     int        Optimistic concurrency version
```

## API

| Method | Path | Description |
|---|---|---|
| GET | /health | Liveness check |
| POST | /users | Create a user |
| GET | /users | List all users |
| GET | /users/{id} | Get a user by ID |
| PUT | /users/{id} | Update a user |
| DELETE | /users/{id} | Delete a user |

### Create User

```
POST /users
Content-Type: application/json

{
  "name": "Alice",
  "email": "alice@example.com",
  "status": "active"
}
```

Response `201 Created`:

```json
{
  "id": "user-20260624120000.000000000",
  "name": "Alice",
  "email": "alice@example.com",
  "status": "active",
  "created_at": "2026-06-24T12:00:00Z",
  "updated_at": "2026-06-24T12:00:00Z",
  "version": 1
}
```

### Update User

Requires the current `version` value. Returns `409 Conflict` on mismatch.

```
PUT /users/{id}
Content-Type: application/json

{
  "name": "Alice Updated",
  "email": "alice@example.com",
  "status": "inactive",
  "version": 1
}
```

Response `200 OK` — returns the updated user with `version` incremented.

### Error Responses

```json
{ "error": "user not found" }
```

| Status | Condition |
|---|---|
| 400 | Missing required fields or malformed body |
| 404 | User does not exist |
| 409 | Version mismatch on update |
| 500 | Internal server error |

## Configuration

Loaded from `config.yaml`. Sensitive values override via environment variables.

```yaml
service:
  name: user-service
  port: 8080
  env: development

log:
  level: info

database:
  provider: postgres
  host: localhost
  port: 5432
  name: userdb
  user: postgres
  password: ""       # override: DATABASE_PASSWORD

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
    name: user.events.main
    routing_key: "user.*"
    durable: true
```

## Running Locally

```bash
cd services/user-service
make run
# or
go run ./cmd
```

Server starts on `:8080`.

## Running Tests

```bash
make test
# or
go test ./...
```

## Docker

Build context is `services/` to include the shared `pkg` module via the Go workspace:

```bash
make docker-build        # builds image: user-service:latest
make docker-run          # runs on port 8080
```

Or manually from `services/`:

```bash
docker build -t user-service:latest -f user-service/docker/Dockerfile .
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

## Dependencies

| Package | Purpose |
|---|---|
| `pkg/config` | YAML + env var configuration |
| `pkg/logger` | Structured JSON logger (zap) |
| `pkg/database` | PostgreSQL abstraction (pgx) |
| `pkg/queue` | RabbitMQ publisher abstraction |
| `go-chi/chi` | HTTP router |

## Events Published

| Event | Trigger |
|---|---|
| `user.created` | Successful user creation |
| `user.updated` | Successful user update |
| `user.deleted` | Successful user deletion |

Event payload:

```json
{
  "user_id": "user-20260624120000.000000000",
  "email": "alice@example.com",
  "event_type": "user.created",
  "timestamp": "2026-06-24T12:00:00Z"
}
```
