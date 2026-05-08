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
  base_tags = {
    App         = var.app_name
    Environment = var.environment
    Owner       = var.owner
    CostCenter  = var.cost_center
    ManagedBy   = "terraform"
  }
}

############################################
# GitHub OIDC IAM Role
############################################

module "github_oidc" {
  source = "./modules/github-oidc-role"

  role_name    = "github-oidc-deploy-role"
  github_owner = "ajawesley"
  github_repo  = "healthcare-event-stream-platform"
  github_ref   = "*"

  inline_policy_statements = [
    {
      Effect = "Allow"
      Action = [
        "sts:AssumeRole",
        "ecr:*",
        "ecs:*",
        "elasticloadbalancing:*",
        "ec2:*",
        "lambda:*",
        "s3:*",
        "glue:*",
        "cloudwatch:*",
        "logs:*",
        "iam:PassRole"
      ]
      Resource = "*"
    }
  ]
}

############################################
# KMS Key (S3 Encryption)
############################################

resource "aws_kms_key" "this" {
  description             = "KMS key for S3 encryption"
  deletion_window_in_days = 7
  enable_key_rotation     = true
  tags                    = local.base_tags
}

############################################
# Raw Events S3 Bucket
############################################

resource "aws_s3_bucket" "this" {
  bucket = var.bucket_name
  tags   = local.base_tags
}

resource "aws_s3_bucket_versioning" "this" {
  bucket = aws_s3_bucket.this.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "this" {
  bucket = aws_s3_bucket.this.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.this.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "this" {
  bucket = aws_s3_bucket.this.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_policy" "this" {
  bucket = aws_s3_bucket.this.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "DenyUnencryptedUploads"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:PutObject"
        Resource  = "${aws_s3_bucket.this.arn}/*"
        Condition = {
          StringNotEquals = {
            "s3:x-amz-server-side-encryption" = "aws:kms"
          }
        }
      }
    ]
  })
}

############################################
# Access Logs Bucket (ALB)
############################################

resource "aws_s3_bucket" "access_logs" {
  bucket = var.access_log_bucket_name
  tags   = local.base_tags
}

############################################
# CloudWatch Log Group (ECS)
############################################

resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/${var.app_name}/${var.environment}/ecs"
  retention_in_days = 30
  tags              = local.base_tags
}

############################################
# ECS Cluster
############################################

resource "aws_ecs_cluster" "cluster" {
  name = "${var.app_name}-${var.environment}-cluster"
  tags = local.base_tags
}

############################################
# ALB Security Group
############################################

resource "aws_security_group" "alb" {
  name        = "${var.app_name}-${var.environment}-alb-sg"
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

  tags = local.base_tags
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

  tags = local.base_tags
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

  tags = local.base_tags
}

############################################
# ALB Module (HTTP only)
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
  tags        = local.base_tags
}

############################################
# IAM Module
############################################

module "iam" {
  source = "./modules/iam"

  environment = var.environment
  owner       = var.owner
  cost_center = var.cost_center

  raw_bucket_arn    = aws_s3_bucket.this.arn
  script_bucket_arn = "arn:aws:s3:::${var.script_bucket}"
  golden_bucket_arn = "arn:aws:s3:::${var.app_name}-${var.environment}-golden-events-001"

  kms_key_arn   = aws_kms_key.this.arn
  log_group_arn = aws_cloudwatch_log_group.ecs.arn
  tags          = local.base_tags
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
  s3_bucket_name          = aws_s3_bucket.this.bucket
  kms_key_arn             = aws_kms_key.this.arn
  s3_prefix               = "events"
  log_group_name          = aws_cloudwatch_log_group.ecs.name
  desired_count           = var.desired_count
  target_group_arn        = module.alb.target_group_arn
  owner                   = var.owner
  cost_center             = var.cost_center
  tags                    = local.base_tags

  enable_adot      = true
  adot_image       = "public.ecr.aws/aws-observability/aws-otel-collector:latest"
  adot_config_file = "${path.module}/otel/collector-config.yaml"

  dd_api_key        = var.dd_api_key
  honeycomb_api_key = var.honeycomb_api_key
  honeycomb_dataset = var.honeycomb_dataset
}

############################################
# Glue Job Module
############################################

module "glue_job" {
  source = "./modules/glue_job"

  app_name    = var.app_name
  environment = var.environment
  owner       = var.owner
  cost_center = var.cost_center

  glue_role_arn = module.iam.glue_role_arn

  script_bucket  = var.script_bucket
  script_s3_path = var.glue_script_s3_path
  temp_dir       = var.glue_temp_dir

  raw_bucket    = aws_s3_bucket.this.bucket
  golden_bucket = "${var.app_name}-${var.environment}-golden-events-001"

  log_group_name = "/aws/glue/${var.app_name}-${var.environment}"
  tags           = local.base_tags
}

############################################
# Glue Crawlers Module (NEW)
############################################

module "glue_crawlers" {
  source      = "./modules/glue_crawlers"
  environment = var.environment
  tags        = local.base_tags

  events_bucket = "${var.app_name}-${var.environment}-golden-events-001"
  errors_bucket = "${var.app_name}-${var.environment}-golden-events-001/errors"
}

############################################
# Lambda Build
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

  app_name    = var.app_name
  environment = var.environment
  owner       = var.owner
  cost_center = var.cost_center

  glue_job_name = module.glue_job.glue_job_name
  glue_job_arn  = module.glue_job.glue_job_arn

  raw_bucket_name = aws_s3_bucket.this.bucket

  lambda_role_arn  = module.iam.lambda_role_arn
  lambda_role_name = module.iam.lambda_role_name

  lambda_zip_path = "${path.root}/../cmd/lambda/lambda.zip"

  kms_key_arn = aws_kms_key.this.arn
  tags        = local.base_tags

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
  source_arn    = aws_s3_bucket.this.arn
}

resource "aws_s3_bucket_notification" "raw_events_trigger" {
  bucket = aws_s3_bucket.this.bucket

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
