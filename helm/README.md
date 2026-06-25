# Helm Charts

Helm charts for deploying the Cloud-Native User Platform onto Kubernetes. Each service has its own independent chart with separate values files per environment.

---

## Charts

| Chart | Description |
|---|---|
| `user-service/` | REST API — Deployment, Service, Ingress, ConfigMap, Secret, HPA, PDB |
| `notification-service/` | Queue consumer — Deployment, ConfigMap, Secret, HPA, PDB |

> `notification-service` has no `Service` or `Ingress` — it is a queue consumer with no inbound HTTP traffic.

---

## Structure

```text
helm/
├── user-service/
│   ├── Chart.yaml
│   ├── values.yaml              base defaults (all environments)
│   ├── values-dev.yaml          dev overrides
│   ├── values-prod.yaml         prod overrides
│   └── templates/
│       ├── _helpers.tpl         reusable label and name functions
│       ├── deployment.yaml
│       ├── service.yaml
│       ├── configmap.yaml       generates config.yaml mounted into container
│       ├── secret.yaml          injects passwords as env vars
│       ├── ingress.yaml         ALB ingress (enabled in prod only)
│       ├── hpa.yaml             horizontal pod autoscaler (enabled flag)
│       └── pdb.yaml             pod disruption budget (enabled flag)
└── notification-service/
    ├── Chart.yaml
    ├── values.yaml
    ├── values-dev.yaml
    ├── values-prod.yaml
    └── templates/
        ├── _helpers.tpl
        ├── deployment.yaml
        ├── configmap.yaml
        ├── secret.yaml
        ├── hpa.yaml
        └── pdb.yaml
```

---

## How Values Work

`values.yaml` holds all defaults. Environment files only override what differs. Helm deep-merges them at deploy time — you only specify what changes per environment.

```text
values.yaml          ← base: image, config structure, disabled flags
values-dev.yaml      ← override: debug logging, low resources, HPA off
values-prod.yaml     ← override: 3 replicas, high resources, HPA on, Ingress on
```

---

## Relationship: Terraform → Helm → Service

This is how the three layers connect end to end.

```text
┌─────────────────────────────────────────────────────────────────┐
│  Terraform                                                       │
│                                                                  │
│  modules/rds   → output: rds_endpoint                           │
│  modules/ecr   → output: ecr_user_service_url                   │
│  modules/eks   → output: cluster_name                           │
└──────────────────────────┬──────────────────────────────────────┘
                           │  terraform output -raw rds_endpoint
                           │  terraform output -raw ecr_user_service_url
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  Helm                                                            │
│                                                                  │
│  --set config.database.host=$RDS_ENDPOINT                       │
│  --set image.registry=$ECR_URL                                  │
│  --set secrets.databasePassword=$DB_PASSWORD                    │
│                                                                  │
│  ConfigMap  → generates /config.yaml  → mounted into container  │
│  Secret     → DATABASE_PASSWORD env var → injected into pod     │
└──────────────────────────┬──────────────────────────────────────┘
                           │  Kubernetes mounts file + env vars
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  Service (Go)                                                    │
│                                                                  │
│  config.Load("config.yaml")     ← reads /config.yaml (ConfigMap)│
│  v.AutomaticEnv()               ← reads DATABASE_PASSWORD (Secret)│
│                                                                  │
│  cfg.Database.Host     = value from ConfigMap                   │
│  cfg.Database.Password = value from Secret env var              │
│                                                                  │
│  No code changes per environment. Zero awareness of Kubernetes. │
└─────────────────────────────────────────────────────────────────┘
```

### Why passwords are not in the ConfigMap

ConfigMaps are not encrypted — they are readable by anyone with cluster access. Passwords live only in Kubernetes Secrets, injected as environment variables. Viper's `AutomaticEnv()` with `SetEnvKeyReplacer(".", "_")` maps:

```text
database.password  →  DATABASE_PASSWORD  (from Secret)
queue.password     →  QUEUE_PASSWORD     (from Secret)
```

The `config.yaml` in the ConfigMap has `password: ""` — the empty value is always overridden by the env var at runtime.

### Why config changes trigger rolling restarts

The Deployment includes checksum annotations:

```yaml
annotations:
  checksum/config: {{ include .../configmap.yaml | sha256sum }}
  checksum/secret: {{ include .../secret.yaml   | sha256sum }}
```

Any change to values that affects the ConfigMap or Secret changes the checksum, which changes the pod spec, which triggers a rolling restart automatically. No manual restarts needed.

---

## Feature Flags

HPA, PDB, and Ingress are disabled by default. Enable per environment in the values file — no template changes needed.

| Feature | Default | Dev | Prod |
|---|---|---|---|
| HPA | off | off | on |
| PDB | off | off | on |
| Ingress | off | off | on |

---

## Deploy Commands

### Dev

```bash
helm upgrade --install user-service ./helm/user-service \
  -f helm/user-service/values-dev.yaml \
  --set image.registry=$ECR_USER_SERVICE_URL \
  --set image.tag=$GIT_SHA \
  --set config.database.host=$RDS_ENDPOINT \
  --set secrets.databasePassword=$DB_PASSWORD \
  --set secrets.queuePassword=$QUEUE_PASSWORD

helm upgrade --install notification-service ./helm/notification-service \
  -f helm/notification-service/values-dev.yaml \
  --set image.registry=$ECR_NOTIFICATION_URL \
  --set image.tag=$GIT_SHA \
  --set config.queue.host=$RABBITMQ_HOST \
  --set secrets.queuePassword=$QUEUE_PASSWORD
```

### Prod

```bash
helm upgrade --install user-service ./helm/user-service \
  -f helm/user-service/values-prod.yaml \
  --set image.registry=$ECR_USER_SERVICE_URL \
  --set image.tag=$GIT_SHA \
  --set config.database.host=$RDS_ENDPOINT \
  --set ingress.host=api.example.com \
  --set secrets.databasePassword=$DB_PASSWORD \
  --set secrets.queuePassword=$QUEUE_PASSWORD

helm upgrade --install notification-service ./helm/notification-service \
  -f helm/notification-service/values-prod.yaml \
  --set image.registry=$ECR_NOTIFICATION_URL \
  --set image.tag=$GIT_SHA \
  --set config.queue.host=$RABBITMQ_HOST \
  --set secrets.queuePassword=$QUEUE_PASSWORD
```

### Where `$ECR_URL`, `$RDS_ENDPOINT` come from

```bash
cd terraform/environments/dev

RDS_ENDPOINT=$(terraform output -raw rds_endpoint)
ECR_USER_SERVICE_URL=$(terraform output -raw ecr_user_service_url)
ECR_NOTIFICATION_URL=$(terraform output -raw ecr_notification_url)
CLUSTER=$(terraform output -raw cluster_name)

aws eks update-kubeconfig --region eu-west-1 --name $CLUSTER
```

In CI/CD (GitHub Actions) these are captured automatically as part of the deploy job. See `.github/workflows/deploy.yml`.

---

## Validate Before Deploy

```bash
# Lint chart
helm lint ./helm/user-service

# Preview rendered templates without deploying
helm template user-service ./helm/user-service \
  -f helm/user-service/values-prod.yaml \
  --set config.database.host=example.rds.amazonaws.com \
  --set secrets.databasePassword=test
```

---

## Rollback

```bash
helm history user-service               # list releases
helm rollback user-service 2            # roll back to release 2
```
