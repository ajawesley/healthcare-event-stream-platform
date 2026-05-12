############################################
# Providers
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

############################################
# Locals
############################################

locals {
  base_tags = {
    App         = var.app_name
    Environment = var.environment
    Owner       = var.owner
    CostCenter  = var.cost_center
  }
}

############################################
# VPC
############################################

module "vpc" {
  source = "../../modules/vpc"

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
# KMS Key
############################################

resource "aws_kms_key" "this" {
  description             = "KMS key for ${var.app_name}-${var.environment}"
  deletion_window_in_days = 7
  enable_key_rotation     = true
  tags                    = local.base_tags
}

############################################
# S3 Buckets (Raw, Golden, Scripts, Access Logs)
############################################

module "s3_buckets" {
  source = "../../modules/s3_buckets"

  app_name    = var.app_name
  environment = var.environment
  aws_region  = var.aws_region

  owner       = var.owner
  cost_center = var.cost_center

  raw_bucket_name         = var.raw_bucket_name
  access_logs_bucket_name = var.access_logs_bucket_name
  golden_bucket_name      = "${var.app_name}-${var.environment}-golden-events-001"
  script_bucket_name      = "hesp-${var.environment}-glue-scripts-001"

  s3_output_base_path = var.s3_output_base_path
  s3_error_path       = var.s3_error_path
  error_prefix        = "errors"

  glue_temp_dir        = var.glue_temp_dir
  glue_script_s3_path  = var.glue_script_s3_path
  lambda_zip_path      = var.lambda_zip_path

  kms_key_arn = aws_kms_key.this.arn
  tags        = local.base_tags
}

############################################
# DynamoDB (Compliance Rules)
############################################

module "compliance_dynamodb" {
  source = "../../modules/dynamodb_compliance_rules"

  table_name         = "${var.app_name}-${var.environment}-compliance-rules"
  ttl_enabled        = false
  ttl_attribute_name = "expires_at"

  tags = local.base_tags
}

############################################
# IAM
############################################

module "iam" {
  source = "../../modules/iam"

  app_name    = var.app_name
  environment = var.environment
  owner       = var.owner
  cost_center = var.cost_center

  compliance_db_password_secret_arn = var.compliance_db_password_secret_arn

  dynamodb_table_arn = module.compliance_dynamodb.table_arn

  raw_bucket_arn    = module.s3_buckets.raw_bucket_arn
  script_bucket_arn = module.s3_buckets.script_bucket_arn
  golden_bucket_arn = module.s3_buckets.golden_bucket_arn

  kms_key_arn   = aws_kms_key.this.arn
  log_group_arn = aws_cloudwatch_log_group.ecs.arn

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
# ALB
############################################

module "alb" {
  source = "../../modules/alb"

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
# ECS Cluster + Log Group
############################################

resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/${var.app_name}/${var.environment}/ecs"
  retention_in_days = 30
  tags              = local.base_tags
}

resource "aws_ecs_cluster" "cluster" {
  name = "${var.app_name}-${var.environment}-cluster"
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
# Redis (Compliance Cache)
############################################

module "compliance_redis" {
  source = "../../modules/redis_compliance"

  name                    = "${var.app_name}-${var.environment}-compliance-redis"
  vpc_id                  = module.vpc.vpc_id
  isolated_subnet_ids     = module.vpc.isolated_subnets
  ingestion_service_sg_id = aws_security_group.ecs.id

  node_type                  = "cache.t4g.small"
  replicas_per_node_group    = 1
  transit_encryption_enabled = true

  tags = local.base_tags
}

############################################
# RDS (Compliance DB)
############################################

module "compliance_db" {
  source = "../../modules/rds_postgres_compliance_db"

  name                    = "${var.app_name}-${var.environment}-compliance-db"
  vpc_id                  = module.vpc.vpc_id
  isolated_subnet_ids     = module.vpc.isolated_subnets
  ingestion_service_sg_id = aws_security_group.ecs.id

  db_name     = var.compliance_db_name
  db_username = var.compliance_db_username
  db_password = var.compliance_db_password

  tags = local.base_tags
}

############################################
# ECS Service Module
############################################

module "ecs_service" {
  source = "../../modules/ecs_service"

  app_name                          = var.app_name
  environment                       = var.environment
  cluster_name                      = aws_ecs_cluster.cluster.name
  compliance_db_host                = module.compliance_db.db_host
  compliance_db_port                = 5432
  compliance_db_name                = var.compliance_db_name
  compliance_db_username            = var.compliance_db_username
  compliance_db_password_secret_arn = var.compliance_db_password_secret_arn

  container_image = var.container_image
  adot_image      = var.adot_image

  dynamodb_table_name = module.compliance_dynamodb.table_name

  task_execution_role_arn = module.iam.execution_role_arn
  task_role_arn           = module.iam.task_role_arn

  subnet_ids         = module.vpc.private_subnets
  security_group_ids = [aws_security_group.ecs.id]

  s3_bucket_name = module.s3_buckets.raw_bucket_name
  s3_prefix      = "events"

  kms_key_arn    = aws_kms_key.this.arn
  log_group_name = aws_cloudwatch_log_group.ecs.name

  desired_count = var.desired_count

  redis_primary_endpoint = module.compliance_redis.primary_endpoint
  target_group_arn       = module.alb.target_group_arn

  owner       = var.owner
  cost_center = var.cost_center
  tags        = local.base_tags

  enable_adot = true
}
