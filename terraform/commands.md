cd terraform/environments/dev

# 1. Initialise — downloads providers, connects to S3 backend
terraform init

# 2. Validate syntax
terraform validate

# 3. Preview changes
terraform plan -var="db_password=$DB_PASSWORD"

# 4. Apply
terraform apply -var="db_password=$DB_PASSWORD"

# 5. After apply — configure kubectl
aws eks update-kubeconfig \
  --region eu-west-1 \
  --name $(terraform output -raw cluster_name)

# 6. Verify cluster
kubectl get nodes

# 7. Get values for Helm deploy
terraform output rds_endpoint
terraform output ecr_urls
