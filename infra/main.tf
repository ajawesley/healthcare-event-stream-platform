############################################
# Terraform + Provider
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
# GitHub OIDC IAM Role (NEW)
############################################

module "github_oidc" {
  source = "./modules/github-oidc-role"

  role_name    = "github-oidc-deploy-role"
  github_owner = "ajawesley"
  github_repo  = "healthcare-event-stream-platform"
  github_ref   = "*" # allow all branches

  inline_policy_statements = [
    {
      Effect = "Allow"
      Action = [
        "sts:AssumeRole",
        "lambda:*",
        "s3:*",
        "iam:PassRole",
        "cloudwatch:*",
        "logs:*",
        "glue:*"
      ]
      Resource = "*"
    }
  ]
}

############################################
# VPC Module
############################################

module "vpc" {
  source = "./modules/vpc"

  name       = "${var.app_name}-${var.environment}"
  region     = var.aws_region
  vpc_cidr   = "10.0.0.0/16"
  primary_az = "us-east-1a"

  azs = {
    "us-east-1a" = { index = 0 }
    "us-east-1b" = { index = 1 }
    "us-east-1c" = { index = 2 }
  }

  tags = local.common_tags
}

############################################
# Access Logs Bucket (ALB)
############################################

resource "aws_s3_bucket" "access_logs" {
  bucket = var.access_log_bucket_name
  tags   = local.common_tags
}

############################################
# CloudWatch Log Group (ECS)
############################################

resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/${var.app_name}/${var.environment}/ecs"
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
  name        = "${var.app_name}-${var.environment}-ecs-sg"
  description = "Security group for ECS tasks"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
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
# S3 Module (Raw Bucket)
############################################

module "s3" {
  source = "./modules/s3"

  bucket_name = var.bucket_name
  environment = var.environment
  owner       = var.owner
  cost_center = var.cost_center
  tags        = local.common_tags
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
  log_group_arn = aws_cloudwatch_log_group.ecs.arn
  tags          = local.common_tags
}

############################################
# ALB Security Group
############################################

resource "aws_security_group" "alb" {
  name        = "${var.app_name}-${var.environment}-alb-sg-new"
  description = "ALB security group"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
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
# ALB Module
############################################

module "alb" {
  source = "./modules/alb"

  app_name              = var.app_name
  environment           = var.environment
  vpc_id                = module.vpc.vpc_id
  subnet_ids            = module.vpc.public_subnets
  alb_security_group_id = aws_security_group.alb.id

  owner       = var.owner
  cost_center = var.cost_center
  tags        = local.common_tags
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
  subnet_ids              = module.vpc.private_subnets
  security_group_ids      = [aws_security_group.ecs.id]
  s3_bucket_name          = module.s3.bucket_name
  kms_key_arn             = module.s3.kms_key_arn
  s3_prefix               = "events"
  log_group_name          = aws_cloudwatch_log_group.ecs.name
  desired_count           = var.desired_count
  target_group_arn        = module.alb.target_group_arn
  owner                   = var.owner
  cost_center             = var.cost_center
  tags                    = local.common_tags
}

############################################
# Glue Job Module
############################################

module "glue_job" {
  source = "./modules/glue_job"

  app_name       = var.app_name
  environment    = var.environment
  owner          = var.owner
  cost_center    = var.cost_center
  glue_role_arn  = module.iam.glue_role_arn
  script_s3_path = var.glue_script_s3_path
  temp_dir       = var.glue_temp_dir
  log_group_name = "/aws/glue/${var.app_name}-${var.environment}"
  tags           = local.common_tags
}

############################################
# Lambda Build (Go)
############################################

resource "null_resource" "build_lambda" {
  triggers = {
    src_hash = filemd5("${path.root}/../cmd/lambda/main.go")
  }

  provisioner "local-exec" {
    command = <<EOF
cd ../cmd/lambda
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap main.go
zip lambda.zip bootstrap
EOF
  }
}

############################################
# Lambda Trigger Module
############################################

module "lambda_trigger" {
  source = "./modules/lambda_trigger"

  app_name         = var.app_name
  environment      = var.environment
  owner            = var.owner
  cost_center      = var.cost_center
  glue_job_name    = module.glue_job.glue_job_name
  glue_job_arn     = module.glue_job.glue_job_arn
  raw_bucket_name  = module.s3.bucket_name
  lambda_role_arn  = module.iam.lambda_role_arn
  lambda_role_name = module.iam.lambda_role_name
  lambda_zip_path  = "${path.root}/../cmd/lambda/lambda.zip"
  tags             = local.common_tags
  kms_key_arn      = module.s3.kms_key_arn

  depends_on = [null_resource.build_lambda]
}

############################################
# S3 → Lambda Notification
############################################

resource "aws_lambda_permission" "s3_invoke" {
  statement_id  = "AllowS3Invoke"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_trigger.lambda_name
  principal     = "s3.amazonaws.com"
  source_arn    = module.s3.bucket_arn
}

resource "aws_s3_bucket_notification" "raw_events_trigger" {
  bucket = module.s3.bucket_name

  lambda_function {
    lambda_function_arn = module.lambda_trigger.lambda_arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "events/"
  }

  depends_on = [
    module.lambda_trigger,
    aws_lambda_permission.s3_invoke
  ]
}
