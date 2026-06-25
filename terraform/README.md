# Terraform Infrastructure

Infrastructure as Code for the Cloud-Native User Platform on AWS. All cloud resources are defined as reusable modules and composed per environment (dev, prod).

## Structure

```text
terraform/
├── environments/
│   ├── dev/          ← dev environment root module
│   └── prod/         ← prod environment root module
└── modules/
    ├── vpc/          ← networking foundation
    ├── eks/          ← Kubernetes cluster
    ├── rds/          ← PostgreSQL database
    ├── ecr/          ← container image registry
    ├── alb/          ← application load balancer
    └── iam/          ← roles and policies
```

Each environment consumes the modules with environment-specific variable values. No module contains hardcoded environment names or account IDs.

---

## AWS Resources

### VPC (`modules/vpc`)

**What it creates:**
- VPC with a defined CIDR block
- Public subnets (one per AZ) — for the ALB
- Private subnets (one per AZ) — for EKS nodes and RDS
- Internet Gateway — outbound access from public subnets
- NAT Gateway — outbound access from private subnets (EKS nodes pulling images)
- Route tables for public and private subnets

**Why:**
EKS nodes and RDS run in private subnets — they are not reachable from the internet directly. Only the ALB sits in public subnets. This is the standard AWS VPC design for production workloads.

**Used by:** `eks`, `rds`, `alb`

---

### EKS (`modules/eks`)

**What it creates:**
- EKS cluster (control plane managed by AWS)
- Managed node group (EC2 worker nodes in private subnets)
- OIDC provider — enables IAM Roles for Service Accounts (IRSA)
- AWS Load Balancer Controller — watches Kubernetes Ingress resources and provisions ALBs
- EKS add-ons: CoreDNS, kube-proxy, VPC CNI

**Why:**
EKS removes the burden of managing the Kubernetes control plane. IRSA allows pods to assume IAM roles without storing credentials — the preferred AWS-native auth pattern.

**Used by:** `alb` (Load Balancer Controller runs on EKS), `iam` (OIDC trust policy)

**Depends on:** `vpc` (private subnets for nodes)

---

### RDS (`modules/rds`)

**What it creates:**
- RDS PostgreSQL instance in a private subnet group
- Security group — allows inbound `5432` only from EKS node security group
- Parameter group — PostgreSQL tuning (connection limits, timeouts)
- Subnet group — spans multiple AZs for failover

**Why:**
RDS is managed PostgreSQL — automated backups, patching, and multi-AZ failover without operational overhead. The security group ensures only application pods can reach the database; it is never exposed to the internet.

**Used by:** user-service (connection details injected via Helm → Kubernetes Secret)

**Depends on:** `vpc` (private subnets)

---

### ECR (`modules/ecr`)

**What it creates:**
- ECR repository for `user-service`
- ECR repository for `notification-service`
- Lifecycle policy — retains the last N images, deletes older ones automatically
- Repository policy — allows EKS node IAM role to pull images

**Why:**
ECR is the private container registry. CI/CD pushes built images here; EKS pulls from here at deploy time. Image tags (e.g. git SHA) are passed to Helm at deploy time to control which version runs.

**Used by:** EKS nodes (image pull), GitHub Actions (image push)

---

### ALB (`modules/alb`)

**What it creates:**
- Security group for the ALB — allows inbound `443` (HTTPS) from the internet
- IAM role for the AWS Load Balancer Controller (via IRSA)
- ACM certificate for TLS termination
- The actual ALB is provisioned dynamically by the Load Balancer Controller when a Kubernetes Ingress resource is created by Helm

**Why:**
The ALB is the single internet-facing entry point for the user-service REST API. TLS is terminated at the ALB — traffic inside the VPC between the ALB and pods runs over HTTP. The controller pattern means Terraform does not manage the ALB listener rules directly — Helm annotations on the Ingress resource drive the ALB configuration.

**Depends on:** `vpc` (public subnets), `eks` (OIDC for IAM role), `iam`

---

### IAM (`modules/iam`)

**What it creates:**
- EKS node IAM role — allows EC2 nodes to join the cluster and pull from ECR
- AWS Load Balancer Controller IAM role — allows the controller to manage ALBs (via IRSA)
- RDS access policy — attached to the node role (for future IAM database auth)
- ECR read policy — attached to the node role

**Why:**
Least-privilege access. Each component gets only the permissions it needs. IRSA scopes IAM roles to specific Kubernetes service accounts rather than granting broad permissions to all pods on a node.

**Used by:** `eks`, `alb`

---

## Provisioning Order

Terraform handles dependency resolution automatically via `depends_on` and resource references, but the logical order is:

```text
1. iam        ← roles needed by everything else
2. vpc        ← networking needed by eks, rds, alb
3. ecr        ← registry needed before images are pushed
4. rds        ← database (independent of eks)
5. eks        ← cluster (needs vpc, iam)
6. alb        ← load balancer controller (needs eks, vpc, iam)
```

---

## Environments

Each environment has its own root module that calls the shared modules with environment-specific values.

```text
environments/dev/
├── main.tf        ← module calls
├── variables.tf   ← input declarations
├── outputs.tf     ← cluster name, RDS endpoint, ECR URLs
└── terraform.tfvars
```

| Variable | dev | prod |
|---|---|---|
| EKS node instance type | `t3.medium` | `t3.large` |
| EKS node count | 2 | 3–10 (auto-scaling) |
| RDS instance class | `db.t3.micro` | `db.t3.medium` |
| RDS multi-AZ | false | true |
| RDS deletion protection | false | true |
| ECR image retention | 5 | 30 |

---

## Prerequisites

Run these steps **once** before any `terraform` command:

**1. Install tools**

```bash
terraform --version     # >= 1.6
aws --version           # AWS CLI v2
kubectl version
helm version
```

**2. Configure AWS credentials**

```bash
aws configure
aws sts get-caller-identity   # verify authentication
```

**3. Create the S3 bucket for remote state**

```bash
aws s3api create-bucket \
  --bucket user-platform-tfstate \
  --region eu-west-1 \
  --create-bucket-configuration LocationConstraint=eu-west-1

aws s3api put-bucket-versioning \
  --bucket user-platform-tfstate \
  --versioning-configuration Status=Enabled

aws s3api put-public-access-block \
  --bucket user-platform-tfstate \
  --public-access-block-configuration \
    BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
```

Versioning is enabled so a previous state can be recovered if the current one is corrupted.

---

## State Management

Terraform state is stored remotely in S3. Each environment uses a separate state key. State is never stored locally or committed to git.

```text
s3://user-platform-tfstate/dev/terraform.tfstate
s3://user-platform-tfstate/prod/terraform.tfstate
```

**Current locking approach — native S3 locking (`use_lockfile = true`)**

Terraform >= 1.10 supports state locking directly via S3. When an `apply` starts, Terraform writes a `.tflock` file to the same bucket. A second concurrent `apply` sees the lock file and waits. No additional AWS resource is needed.

```hcl
terraform {
  backend "s3" {
    bucket       = "user-platform-tfstate"
    key          = "dev/terraform.tfstate"
    region       = "eu-west-1"
    use_lockfile = true
    encrypt      = true
  }
}
```

**TODO: Migrate to DynamoDB locking for older Terraform compatibility**

`use_lockfile = true` requires Terraform >= 1.10. If the project needs to support older Terraform versions (e.g. in CI/CD pipelines pinned to an earlier version), replace `use_lockfile` with DynamoDB locking:

```bash
# Create the DynamoDB lock table
aws dynamodb create-table \
  --table-name user-platform-tflock \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST
```

```hcl
# backend.tf — DynamoDB approach
terraform {
  backend "s3" {
    bucket         = "user-platform-tfstate"
    key            = "dev/terraform.tfstate"
    region         = "eu-west-1"
    dynamodb_table = "user-platform-tflock"
    encrypt        = true
  }
}
```

Both approaches provide the same protection — only one `apply` can run at a time per environment.

---

## Usage

```bash
cd terraform/environments/dev

terraform init
terraform plan -var-file="terraform.tfvars"
terraform apply -var-file="terraform.tfvars"
```

After `apply`, outputs provide the values needed for Helm:

```bash
terraform output eks_cluster_name    # → configure kubectl
terraform output rds_endpoint        # → Helm values (database.host)
terraform output ecr_user_service    # → image registry URL
```

---

## How Terraform and Helm Connect

Terraform provisions infrastructure. Helm deploys the application onto it. The hand-off happens via Terraform outputs passed as Helm values:

```bash
helm upgrade --install user-service ./helm/user-service \
  --set config.database.host=$(terraform output -raw rds_endpoint) \
  --set image.registry=$(terraform output -raw ecr_user_service) \
  --set image.tag=$GIT_SHA
```

Secrets (database password, queue password) are stored in AWS Secrets Manager and injected via the External Secrets Operator — never passed as plain Helm values.
