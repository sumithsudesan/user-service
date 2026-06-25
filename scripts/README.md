# Scripts

Utility scripts for deploying and managing the Cloud-Native User Platform.

---

## `deploy.sh`

End-to-end deploy script. Reads Terraform outputs and passes them directly to Helm as `--set` flags. This is the local equivalent of what the GitHub Actions deploy workflow does in CI/CD.

### What it does

```text
1. Read Terraform outputs (RDS endpoint, ECR URLs, cluster name)
2. Configure kubectl to target the EKS cluster
3. helm upgrade --install user-service        (with env-specific values)
4. helm upgrade --install notification-service (with env-specific values)
```

### Prerequisites

```bash
# Must be authenticated to AWS
aws sts get-caller-identity

# Must have Terraform state accessible (S3 backend)
cd terraform/environments/dev && terraform output

# Required env vars
export DB_PASSWORD=...
export QUEUE_PASSWORD=...
export RABBITMQ_HOST=...
export GIT_SHA=...        # image tag to deploy (defaults to "latest")
```

### Usage

```bash
# Deploy to dev (default)
./scripts/deploy.sh

# Deploy to prod
./scripts/deploy.sh prod
```

### How Terraform and Helm connect

The script is the explicit hand-off point between infrastructure and application deployment:

```text
terraform output -raw rds_endpoint          → --set config.database.host
terraform output -raw ecr_user_service_url  → --set image.registry (user-service)
terraform output -raw ecr_notification_url  → --set image.registry (notification-service)
terraform output -raw cluster_name          → aws eks update-kubeconfig
```

Passwords (`DB_PASSWORD`, `QUEUE_PASSWORD`) are never stored in Terraform or Helm values files — they are passed at deploy time from environment variables or a secrets manager.

---

## `db/001_create_users.sql`

Database migration — creates the `users` table in PostgreSQL.

Run once after the RDS instance is provisioned by Terraform:

```bash
psql -h $RDS_ENDPOINT -U postgres -d userdb -f scripts/db/001_create_users.sql
```
