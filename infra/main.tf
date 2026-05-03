############################################
# Provider
############################################

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

############################################
# Common Tags
############################################

locals {
  common_tags = {
    App         = var.app_name
    Environment = var.environment
    Owner       = var.owner
    CostCenter  = var.cost_center
    ManagedBy   = "terraform"
  }
}

############################################
# Central Access Logs Bucket
############################################

resource "aws_s3_bucket" "access_logs" {
  bucket = var.access_log_bucket_name
  tags   = local.common_tags
}

resource "aws_s3_bucket_policy" "access_logs" {
  bucket = aws_s3_bucket.access_logs.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AWSALBAccessLogs"
        Effect = "Allow"
        Principal = {
          Service = "logdelivery.elasticloadbalancing.amazonaws.com"
        }
        Action   = ["s3:PutObject"]
        Resource = "${aws_s3_bucket.access_logs.arn}/*"
      }
    ]
  })
}

############################################
# CloudWatch Log Group
############################################

resource "aws_cloudwatch_log_group" "ingest" {
  name              = "/hesp/${var.environment}/ingest"
  retention_in_days = 30
  tags              = local.common_tags
}

############################################
# ECS Cluster
############################################

resource "aws_ecs_cluster" "cluster" {
  name = "${var.app_name}-${var.environment}-cluster"
  tags = local.common_tags
}

############################################
# ECS Security Group
############################################

resource "aws_security_group" "ecs" {
  name        = "${var.app_name}-ecs-sg"
  description = "Security group for ECS ingest tasks"
  vpc_id      = var.vpc_id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [module.alb.alb_security_group_id]
    description     = "Allow ALB to reach ECS tasks on port 8080"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.common_tags
}

############################################
# S3 Module
############################################

module "s3" {
  source = "./modules/s3"

  bucket_name          = var.bucket_name
  environment          = var.environment
  owner                = var.owner
  cost_center          = var.cost_center
  ingest_task_role_arn = module.iam.task_role_arn
  access_log_bucket_id = aws_s3_bucket.access_logs.id

  tags = local.common_tags
}

############################################
# IAM Module
############################################

module "iam" {
  source = "./modules/iam"

  environment   = var.environment
  owner         = var.owner
  cost_center   = var.cost_center
  bucket_arn    = module.s3.bucket_arn
  kms_key_arn   = module.s3.kms_key_arn
  log_group_arn = aws_cloudwatch_log_group.ingest.arn

  tags = local.common_tags
}

############################################
# ALB Module (HTTP-only)
############################################

module "alb" {
  source = "./modules/alb"

  app_name              = var.app_name
  environment           = var.environment
  vpc_id                = var.vpc_id
  subnet_ids            = var.public_subnet_ids
  ecs_security_group_id = aws_security_group.ecs.id
  access_log_bucket_id  = aws_s3_bucket.access_logs.id
  owner                 = var.owner
  cost_center           = var.cost_center

  tags = local.common_tags
}

############################################
# ECS Service Module
############################################

module "ecs_service" {
  source = "./modules/ecs_service"

  app_name                = var.app_name
  environment             = var.environment
  cluster_name            = aws_ecs_cluster.cluster.name
  container_image         = var.container_image
  task_execution_role_arn = module.iam.execution_role_arn
  task_role_arn           = module.iam.task_role_arn
  subnet_ids              = var.private_subnet_ids
  security_group_ids      = [aws_security_group.ecs.id]
  s3_bucket_name          = module.s3.bucket_name
  log_group_name          = aws_cloudwatch_log_group.ingest.name
  desired_count           = var.desired_count
  target_group_arn        = module.alb.target_group_arn
  owner                   = var.owner
  cost_center             = var.cost_center

  tags = local.common_tags
}

############################################
# Outputs
############################################

output "alb_dns_name" {
  value = module.alb.alb_dns_name
}

output "ecs_service_arn" {
  value = module.ecs_service.service_arn
}

output "s3_bucket_name" {
  value = module.s3.bucket_name
}
