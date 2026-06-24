# User Service

The user service is the core domain service of the Cloud-Native User Platform. It provides a REST API for managing user lifecycle operations and is designed around clean architecture principles — business logic is decoupled from infrastructure concerns via interface-driven design.

## Responsibilities

- Create, read, update, and delete users
- Enforce domain validation rules
- Expose structured error responses
- Support optimistic concurrency control via versioning

## Architecture

```text
HTTP Request
    │
    ▼
Router (chi)
    │
    ▼
Handler (HTTP layer — request parsing, response serialisation)
    │
    ▼
Service (domain logic — validation, business rules)
    │
    ▼
Storage (in-memory map — temporary, pending PostgreSQL)
```

The handler layer never contains business logic. The service layer never contains HTTP concerns. This separation makes each layer independently testable and replaceable.

## Domain Model

```text
User
├── id          string     Unique identifier
├── name        string     Display name
├── email       string     Contact email
├── status      string     Account status (e.g. active, inactive)
├── created_at  time.Time  UTC creation timestamp
├── updated_at  time.Time  UTC last-modified timestamp
└── version     int        Optimistic concurrency version
```

## API

| Method | Path          | Description         |
|--------|---------------|---------------------|
| GET    | /health       | Liveness check      |
| POST   | /users        | Create a user       |
| GET    | /users        | List all users      |
| GET    | /users/{id}   | Get a user by ID    |
| PUT    | /users/{id}   | Update a user       |
| DELETE | /users/{id}   | Delete a user       |

### Create User

```
POST /users
```

Request:

```json
{
  "name": "test user",
  "email": "test user.2026@gmail.com",
  "status": "active"
}
```

Response `201 Created`:

```json
{
  "id": "user-20260624120000.000000000",
  "name": "test user",
  "email": "test user.2026@gmail.com",
  "status": "active",
  "created_at": "2026-06-24T12:00:00Z",
  "updated_at": "2026-06-24T12:00:00Z",
  "version": 1
}
```

### Update User

Update uses optimistic concurrency control. The `version` field in the request must match the current version of the record. If it does not match, the request is rejected with `409 Conflict`.

```
PUT /users/{id}
```

Request:

```json
{
  "name": "test user",
  "email": "test user.2026@gmail.com",
  "status": "inactive",
  "version": 1
}
```

Response `200 OK`: returns the updated user with `version` incremented.

### Error Responses

All errors return a JSON body:

```json
{
  "error": "user not found"
}
```

| Status | Condition                        |
|--------|----------------------------------|
| 400    | Missing required fields          |
| 400    | Malformed request body           |
| 404    | User ID does not exist           |
| 409    | Version mismatch on update       |
| 500    | Internal server error            |

## Running Locally

```bash
cd services/user-service
go run ./cmd/main.go
```

Server starts on `:8080`.

## Running Tests

```bash
cd services/user-service
go test ./tests/...
```

## Dependencies

| Package            | Purpose          |
|--------------------|------------------|
| `go-chi/chi/v5`    | HTTP router      |

## Current Limitations

This implementation uses an **in-memory map** for storage. Data is lost on restart. This is intentional as a temporary layer while the PostgreSQL repository is being built.

Planned replacements per the project roadmap:

- Storage: in-memory → PostgreSQL
- Events: none → RabbitMQ publisher
- Config: hardcoded port → YAML / environment variables
- Logging: `log` → structured JSON logger
