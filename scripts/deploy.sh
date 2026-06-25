#!/bin/bash
set -e

ENV=${1:-dev}

cd terraform/environments/$ENV
RDS_ENDPOINT=$(terraform output -raw rds_endpoint)
ECR_USER=$(terraform output -raw ecr_user_service_url)
ECR_NOTIF=$(terraform output -raw ecr_notification_url)
CLUSTER=$(terraform output -raw cluster_name)

aws eks update-kubeconfig --region eu-west-1 --name $CLUSTER

helm upgrade --install user-service ./helm/user-service \
  -f helm/user-service/values-$ENV.yaml \
  --set image.registry=$ECR_USER \
  --set image.tag=${GIT_SHA:-latest} \
  --set config.database.host=$RDS_ENDPOINT \
  --set secrets.databasePassword=$DB_PASSWORD \
  --set secrets.queuePassword=$QUEUE_PASSWORD

helm upgrade --install notification-service ./helm/notification-service \
  -f helm/notification-service/values-$ENV.yaml \
  --set image.registry=$ECR_NOTIF \
  --set image.tag=${GIT_SHA:-latest} \
  --set config.queue.host=$RABBITMQ_HOST \
  --set secrets.queuePassword=$QUEUE_PASSWORD
