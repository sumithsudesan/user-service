# Cloud-Native User Platform

A production-grade microservices platform demonstrating senior-level backend, platform, and cloud engineering practices. The business domain is intentionally simple — user lifecycle management — so the architecture, patterns, and infrastructure decisions remain the focus.

## What This Demonstrates

- Clean Architecture and Domain-Driven Design
- Interface-driven design with dependency inversion
- Event-driven async communication via message queues
- Shared library module (Go workspace)
- Configuration-driven behaviour (YAML + env var overrides)
- Structured JSON logging
- Lightweight multi-stage Docker builds (scratch final image)
- Kubernetes-native deployment (Helm, HPA, PDB)
- Infrastructure as Code (Terraform)
- CI/CD automation (GitHub Actions)

## Services

| Service | Description |
|---|---|
| [user-service](services/user-service/) | REST API for user lifecycle (create, read, update, delete). Publishes domain events to RabbitMQ. |
| [notification-service](services/notification-service/) | Event consumer. Subscribes to user domain events and triggers downstream actions. |
| [pkg](services/pkg/) | Shared library: config, logger, database, and queue abstractions used by all services. |

## Architecture

```text
Client
  │
  ▼
user-service ──── PostgreSQL
  │
  └──── RabbitMQ (user.events)
              │
              ▼
     notification-service
              │
              ▼
       Email / SMS / Webhook  (planned)
```

All services share the `pkg` module via a Go workspace. Business logic depends only on interfaces defined in `pkg` — not on infrastructure implementations. Swapping PostgreSQL for another database, or RabbitMQ for Kafka, requires no changes to business logic.

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.26 |
| HTTP router | chi |
| Database | PostgreSQL (pgx) |
| Message broker | RabbitMQ (amqp091-go) |
| Logger | zap (structured JSON) |
| Config | Viper (YAML + env vars) |
| Container | Docker (scratch final image) |
| Orchestration | Kubernetes + Helm |
| Infrastructure | Terraform |
| CI/CD | GitHub Actions |

## Repository Structure

```text
user-service/
├── docs/                         Architecture diagrams
├── helm/                         Helm chart
├── scripts/                      Utility scripts
├── services/
│   ├── go.work                   Go workspace (links all modules)
│   ├── pkg/                      Shared library (config, logger, database, queue)
│   ├── user-service/             User domain REST API
│   └── notification-service/     Event consumer
├── terraform/                    Infrastructure as Code
└── .github/workflows/            CI/CD pipelines
```

## Quick Start

**Prerequisites:** Go 1.26, Docker, PostgreSQL, RabbitMQ

```bash
# Start infrastructure
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:16-alpine
docker run -d -p 5672:5672 -p 15672:15672 rabbitmq:3-management-alpine

# Run user-service
cd services/user-service
go run ./cmd

# Run notification-service (separate terminal)
cd services/notification-service
go run ./cmd
```

## Running Tests

```bash
cd services/user-service          && go test ./...
cd services/notification-service  && go test ./...
cd services/pkg                   && go test ./...
```

## Docker

```bash
# Build images (run from each service directory)
cd services/user-service          && make docker-build
cd services/notification-service  && make docker-build
```

## Roadmap

| Phase | Focus | Status |
|---|---|---|
| Phase 1 | REST API, PostgreSQL, RabbitMQ, Docker | In progress |
| Phase 2 | Redis cache, metrics, Helm, Kubernetes | Planned |
| Phase 3 | Terraform, GitHub Actions, multi-env deployment | Planned |
| Phase 4 | OpenTelemetry, distributed tracing, Kafka, auth | Planned |

---

**Author:** Sumith Sudhesan
