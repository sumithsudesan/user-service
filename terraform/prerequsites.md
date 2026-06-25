# Prerequisites

## Tools

```bash
terraform --version     # >= 1.6
aws --version           # AWS CLI v2
kubectl version         # verify cluster after apply
helm version            # deploy services after infra is up
```

## AWS Credentials

```bash
aws configure
aws sts get-caller-identity   # verify authentication works
```

## Remote State — S3 Bucket

Terraform stores its state file in S3. Create the bucket once manually before running any Terraform commands.

**Why:** Terraform is stateless — the state file is how it tracks which AWS resources exist. Storing it in S3 means CI/CD pipelines and multiple engineers share the same state. `use_lockfile = true` in `backend.tf` prevents two applies running at the same time using a native S3 lock (no DynamoDB required).

**State file location after apply:**
```
s3://user-platform-tfstate/dev/terraform.tfstate
s3://user-platform-tfstate/prod/terraform.tfstate
```

**Create the bucket:**

```bash
# Create bucket
aws s3api create-bucket \
  --bucket user-platform-tfstate \
  --region eu-west-1 \
  --create-bucket-configuration LocationConstraint=eu-west-1

# Enable versioning — allows state recovery if corrupted
aws s3api put-bucket-versioning \
  --bucket user-platform-tfstate \
  --versioning-configuration Status=Enabled

# Block all public access
aws s3api put-public-access-block \
  --bucket user-platform-tfstate \
  --public-access-block-configuration \
    BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
```

No DynamoDB table needed — locking is handled natively by S3 (`use_lockfile = true` in `backend.tf`).

## Module Write Order

```text
1. modules/iam       ← roles referenced by everything else
2. modules/vpc       ← networking referenced by eks, rds, alb
3. modules/ecr       ← standalone, no dependencies
4. modules/rds       ← needs vpc outputs
5. modules/eks       ← needs vpc + iam outputs
6. modules/alb       ← needs eks + vpc + iam outputs
7. environments/dev  ← wires all modules together
8. environments/prod ← same structure, different values
```

## Apply Order

```bash
cd terraform/environments/dev

terraform init
terraform validate
terraform plan  -var="db_password=$DB_PASSWORD"
terraform apply -var="db_password=$DB_PASSWORD"

# Configure kubectl after cluster is up
aws eks update-kubeconfig \
  --region eu-west-1 \
  --name $(terraform output -raw cluster_name)

kubectl get nodes
```
