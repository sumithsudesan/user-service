terraform {
  required_version = ">= 1.6"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
  default_tags { tags = local.tags }
}

provider "helm" {
  kubernetes {
    host                   = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_ca)
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      args        = ["eks", "get-token", "--cluster-name", local.cluster_name]
      command     = "aws"
    }
  }
}

locals {
  cluster_name = "user-platform-${var.environment}"
  tags = {
    Environment = var.environment
    Project     = "user-platform"
    ManagedBy   = "terraform"
  }
}

module "iam" {
  source            = "../../modules/iam"
  cluster_name      = local.cluster_name
  oidc_provider_arn = module.eks.oidc_provider_arn
  oidc_provider_url = module.eks.oidc_provider_url
  tags              = local.tags
}

module "vpc" {
  source          = "../../modules/vpc"
  name            = "user-platform-${var.environment}"
  cidr            = "10.0.0.0/16"
  azs             = ["eu-west-1a", "eu-west-1b"]
  public_subnets  = ["10.0.1.0/24", "10.0.2.0/24"]
  private_subnets = ["10.0.11.0/24", "10.0.12.0/24"]
  cluster_name    = local.cluster_name
  tags            = local.tags
}

module "ecr" {
  source           = "../../modules/ecr"
  repositories     = ["user-service", "notification-service"]
  image_retention  = 5
  tags             = local.tags
}

module "eks" {
  source               = "../../modules/eks"
  cluster_name         = local.cluster_name
  subnet_ids           = module.vpc.private_subnet_ids
  node_role_arn        = module.iam.eks_node_role_arn
  node_instance_types  = ["t3.medium"]
  node_desired         = 2
  node_min             = 1
  node_max             = 4
  tags                 = local.tags
}

module "rds" {
  source              = "../../modules/rds"
  identifier          = "user-platform-${var.environment}"
  instance_class      = "db.t3.micro"
  db_name             = "userdb"
  username            = "postgres"
  password            = var.db_password
  subnet_ids          = module.vpc.private_subnet_ids
  vpc_id              = module.vpc.vpc_id
  allowed_sg_id       = module.eks.node_sg_id
  multi_az            = false
  deletion_protection = false
  tags                = local.tags
}

module "alb" {
  source        = "../../modules/alb"
  cluster_name  = local.cluster_name
  lbc_role_arn  = module.iam.lbc_role_arn
  tags          = local.tags
}
