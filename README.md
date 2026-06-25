# Cloud-Native User Platform

A production-grade microservices platform demonstrating senior-level backend, platform, and cloud engineering practices. The business domain is intentionally simple — user lifecycle management — so the architecture, patterns, and infrastructure decisions remain the focus.

---

## What This Demonstrates

- Clean Architecture and Domain-Driven Design
- Interface-driven design with dependency inversion
- Event-driven async communication via message queues
- Shared library module (Go workspace)
- Configuration-driven behaviour (YAML + env var overrides)
- Structured JSON logging (zap)
- Lightweight multi-stage Docker builds (scratch final image)
- Kubernetes-native deployment (Helm, HPA, PDB, Ingress)
- Infrastructure as Code (Terraform — VPC, EKS, RDS, ECR, ALB, IAM)
- CI/CD automation (GitHub Actions — CI on push, deploy on manual trigger)

---

## Services

| Service | Description |
|---|---|
| [user-service](services/user-service/) | REST API for user lifecycle (create, read, update, delete). Publishes domain events to RabbitMQ. |
| [notification-service](services/notification-service/) | Event consumer. Subscribes to user domain events and triggers downstream actions. |
| [pkg](services/pkg/) | Shared library: config, logger, database, and queue abstractions used by all services. |

---

## Service Architecture

```text
Client
  │
  ▼
ALB (AWS Application Load Balancer)
  │
  ▼
user-service (Kubernetes — 3 replicas in prod)
  │              │
  │              ▼
  │         PostgreSQL (RDS — private subnet)
  │
  └──── RabbitMQ  (user.events exchange)
              │
              ▼
     notification-service (Kubernetes — 2 replicas in prod)
              │
              ▼
       Email / SMS / Webhook  (planned)
```

Each service reads configuration from a ConfigMap-generated `config.yaml` mounted at `/config.yaml`. Passwords are injected as environment variables from Kubernetes Secrets — never stored in config files. Business logic is unaware of Kubernetes or any specific infrastructure.

---

## Platform Architecture

```text
GitHub
  │
  ├── push → ci.yml       ── go test, docker build, helm lint, terraform validate
  │
  └── manual → deploy.yml ── terraform apply
                              build + push to ECR
                              helm upgrade --install
                              kubectl rollout status

Terraform (environments/dev | prod)
  ├── modules/vpc          ── VPC, public/private subnets, NAT gateway
  ├── modules/eks          ── EKS cluster, node group, OIDC (IRSA)
  ├── modules/rds          ── PostgreSQL (private subnet, SG scoped to EKS)
  ├── modules/ecr          ── Container registry (user-service, notification-service)
  ├── modules/alb          ── AWS Load Balancer Controller (IRSA)
  └── modules/iam          ── Node role, LBC role, least-privilege policies

Terraform outputs → Helm --set flags (RDS endpoint, ECR URLs, cluster name)

Helm (helm/user-service | notification-service)
  ├── values.yaml          ── base defaults
  ├── values-dev.yaml      ── dev overrides (1 replica, debug logging, HPA off)
  ├── values-prod.yaml     ── prod overrides (3 replicas, HPA on, PDB on, Ingress on)
  └── templates/
      ├── configmap.yaml   ── generates config.yaml from values
      ├── secret.yaml      ── injects DATABASE_PASSWORD, QUEUE_PASSWORD as env vars
      ├── deployment.yaml  ── mounts ConfigMap, reads Secret, checksum restart
      ├── ingress.yaml     ── ALB ingress (prod only)
      ├── hpa.yaml         ── autoscaling (prod only)
      └── pdb.yaml         ── availability guarantee during rolling updates
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.26 |
| HTTP router | chi |
| Database | PostgreSQL (pgx) |
| Message broker | RabbitMQ (amqp091-go) |
| Logger | zap (structured JSON) |
| Config | Viper (YAML + env var override) |
| Container | Docker (scratch final image) |
| Orchestration | Kubernetes + Helm |
| Cloud | AWS (EKS, RDS, ECR, ALB, VPC, IAM) |
| Infrastructure | Terraform |
| CI/CD | GitHub Actions |

---

## Repository Structure

```text
user-service/
├── .github/
│   └── workflows/
│       ├── ci.yml                    lint, test, build, validate on every push
│       └── deploy.yml                manual deploy to dev or prod
├── docs/                             architecture diagrams
├── helm/
│   ├── user-service/                 Helm chart — HTTP API service
│   │   ├── values.yaml
│   │   ├── values-dev.yaml
│   │   ├── values-prod.yaml
│   │   └── templates/
│   └── notification-service/         Helm chart — queue consumer
│       ├── values.yaml
│       ├── values-dev.yaml
│       ├── values-prod.yaml
│       └── templates/
├── scripts/
│   ├── deploy.sh                     local deploy script (Terraform → Helm)
│   └── db/                           database migration SQL
├── services/
│   ├── go.work                       Go workspace (links all modules)
│   ├── pkg/                          shared library (config, logger, database, queue)
│   ├── user-service/                 user domain REST API
│   └── notification-service/         event consumer
└── terraform/
    ├── environments/
    │   ├── dev/                      dev root module
    │   └── prod/                     prod root module
    └── modules/
        ├── vpc/  eks/  rds/
        └── ecr/  alb/  iam/
```

---

## Quick Start (Local)

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
cd services/user-service          && make docker-build
cd services/notification-service  && make docker-build
```

## Deploy to AWS

```bash
# One-shot local deploy (reads Terraform outputs, runs Helm)
ENV=dev \
DB_PASSWORD=... \
QUEUE_PASSWORD=... \
RABBITMQ_HOST=... \
./scripts/deploy.sh dev
```

Or trigger manually via GitHub Actions → **Deploy** → Run workflow.

---

## Roadmap

| Phase | Focus | Status |
|---|---|---|
| Phase 1 | REST API, PostgreSQL, RabbitMQ, Docker, pkg shared library | In progress |
| Phase 2 | Helm charts, Terraform infrastructure, GitHub Actions CI/CD | In progress |
| Phase 3 | Redis cache, Prometheus metrics, multi-env Kubernetes deployment | Planned |
| Phase 4 | OpenTelemetry, distributed tracing, Kafka provider, JWT / OAuth2 auth | Planned |

---

**Author:** Sumith Sudhesan
