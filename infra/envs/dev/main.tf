############################################
# Terraform + Providers
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
# VPC (Landing Zone Baseline)
############################################

module "vpc" {
  source = "../../modules/vpc"

  name_prefix = "${var.app_name}-${var.environment}"
  vpc_cidr    = "10.10.0.0/16"
  az_count    = 3

  tags = local.base_tags
}

############################################
# VPC Endpoints (Private Connectivity)
############################################

module "endpoints" {
  source = "../../modules/endpoints"

  name_prefix        = "${var.app_name}-${var.environment}"
  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids
  vpc_cidr           = module.vpc.vpc_cidr
  region             = var.aws_region

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
# S3 Buckets (Raw, Golden, Scripts, Access Logs, Log Archive)
############################################

module "s3_buckets" {
  source = "../../modules/s3_buckets"

  raw_bucket_name         = "${var.app_name}-${var.environment}-raw-events-001"
  golden_bucket_name      = "${var.app_name}-${var.environment}-golden-events-001"
  script_bucket_name      = "${var.app_name}-${var.environment}-glue-scripts-001"
  access_logs_bucket_name = "${var.app_name}-${var.environment}-access-logs-001"

  # Centralized log archive bucket (AWS Config, CloudTrail, GuardDuty)
  log_archive_bucket_name = var.log_archive_bucket_name

  kms_key_arn  = aws_kms_key.this.arn
  error_prefix = "errors"

  tags = local.base_tags
}

############################################
# Glue Job
############################################

module "glue_job" {
  source = "../../modules/glue_job"

  app_name    = var.app_name
  environment = var.environment
  owner       = var.owner
  cost_center = var.cost_center
  tags        = local.base_tags

  glue_role_arn = module.iam.glue_role_arn

  script_bucket  = module.s3_buckets.scripts_bucket_name
  script_s3_path = var.glue_script_s3_path
  temp_dir       = var.glue_temp_dir

  log_group_name = "/aws/glue/${var.app_name}-${var.environment}-job"

  kms_key_arn = aws_kms_key.this.arn

  raw_bucket    = module.s3_buckets.raw_bucket_name
  golden_bucket = module.s3_buckets.golden_bucket_name
}

############################################
# Glue Crawlers
############################################

module "glue_crawlers" {
  source = "../../modules/glue_crawlers"

  environment = var.environment
  tags        = local.base_tags

  events_bucket = module.s3_buckets.raw_bucket_name
  errors_bucket = "${module.s3_buckets.golden_bucket_name}/errors"
}

############################################
# Lambda Trigger for Glue Job
############################################

module "lambda_trigger" {
  source = "../../modules/lambda_trigger"

  app_name    = var.app_name
  environment = var.environment
  owner       = var.owner
  cost_center = var.cost_center
  tags        = local.base_tags

  glue_job_name = module.glue_job.glue_job_name
  glue_job_arn  = module.glue_job.glue_job_arn

  raw_bucket_name = module.s3_buckets.raw_bucket_name

  lambda_role_arn  = module.iam.lambda_role_arn
  lambda_role_name = module.iam.lambda_role_name

  lambda_zip_path = var.lambda_zip_path
  kms_key_arn     = aws_kms_key.this.arn

  output_base_path = var.s3_output_base_path
  error_path       = var.s3_error_path
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
  script_bucket_arn = module.s3_buckets.scripts_bucket_arn
  golden_bucket_arn = module.s3_buckets.golden_bucket_arn

  kms_key_arn   = aws_kms_key.this.arn
  log_group_arn = aws_cloudwatch_log_group.ecs.arn

  # Required for AWS Config Recorder role
  log_archive_bucket_arn = module.s3_buckets.log_archive_bucket_arn

  tags = local.base_tags
}

############################################
# AWS Config (Landing Zone Baseline)
############################################

module "config" {
  source = "../../modules/config"

  name_prefix = "${var.app_name}-${var.environment}"
  region      = var.aws_region

  log_archive_bucket_arn  = module.s3_buckets.log_archive_bucket_arn
  log_archive_bucket_name = module.s3_buckets.log_archive_bucket_name

  kms_key_arn     = aws_kms_key.this.arn
  config_role_arn = module.iam.config_role_arn

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
  subnet_ids            = module.vpc.public_subnet_ids
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
  isolated_subnet_ids     = module.vpc.isolated_subnet_ids
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
  isolated_subnet_ids     = module.vpc.isolated_subnet_ids
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

  subnet_ids         = module.vpc.private_subnet_ids
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
