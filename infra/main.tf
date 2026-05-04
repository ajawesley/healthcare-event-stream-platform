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
    App                 = var.app_name
    Environment         = var.environment
    Owner               = var.owner
    CostCenter          = var.cost_center
    ManagedBy           = "terraform"
    DataClassification  = "phi"
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
# CloudWatch Log Group (ECS)
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
# Raw Events Bucket (S3 Module)
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
# IAM Module (ECS + Glue Roles)
############################################

module "iam" {
  source = "./modules/iam"

  environment        = var.environment
  owner              = var.owner
  cost_center        = var.cost_center
  bucket_arn         = module.s3.bucket_arn
  kms_key_arn        = module.s3.kms_key_arn
  log_group_arn      = aws_cloudwatch_log_group.ingest.arn
  scripts_bucket_arn = module.glue_scripts_bucket.bucket_arn

  tags = local.common_tags
}

############################################
# ALB Module
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
# Glue Scripts Bucket
############################################

module "glue_scripts_bucket" {
  source = "./modules/s3_bucket"

  bucket_name = "${var.app_name}-${var.environment}-glue-scripts"
  environment = var.environment
  owner       = var.owner
  cost_center = var.cost_center
  tags        = local.common_tags
}

############################################
# Upload Glue Job Script + Python Libraries
############################################

resource "aws_s3_object" "glue_main_script" {
  bucket = module.glue_scripts_bucket.bucket_name
  key    = "scripts/glue_job.py"
  source = "${path.module}/glue/job/glue_job.py"
  etag   = filemd5("${path.module}/glue/job/glue_job.py")
}

resource "aws_s3_object" "glue_libs" {
  for_each = fileset("${path.module}/glue/job", "*.py")

  bucket = module.glue_scripts_bucket.bucket_name
  key    = "scripts/lib/${each.value}"
  source = "${path.module}/glue/job/${each.value}"
  etag   = filemd5("${path.module}/glue/job/${each.value}")
}

############################################
# Glue Job Module
############################################

module "glue_job" {
  source = "./modules/glue_job"

  app_name        = var.app_name
  environment     = var.environment
  owner           = var.owner
  cost_center     = var.cost_center
  glue_role_arn   = module.iam.glue_role_arn
  script_s3_path  = "s3://${module.glue_scripts_bucket.bucket_name}/scripts/glue_job.py"
  temp_dir        = "s3://${module.glue_scripts_bucket.bucket_name}/tmp/"
  log_group_name  = "/aws/glue/${var.app_name}-${var.environment}"

  extra_py_files = [
    "s3://${module.glue_scripts_bucket.bucket_name}/scripts/lib/canonical_event_schema.py",
    "s3://${module.glue_scripts_bucket.bucket_name}/scripts/lib/partitioner.py",
    "s3://${module.glue_scripts_bucket.bucket_name}/scripts/lib/writer.py",
    "s3://${module.glue_scripts_bucket.bucket_name}/scripts/lib/error_writer.py",
    "s3://${module.glue_scripts_bucket.bucket_name}/scripts/lib/metrics.py"
  ]

  number_of_workers = 2
  worker_type       = "G.1X"
  timeout_minutes   = 120

  tags = local.common_tags
}

############################################
# Lambda Build (Go)
############################################

resource "null_resource" "build_lambda" {
  triggers = {
    src_hash = filemd5("${path.module}/lambda/main.go")
  }

  provisioner "local-exec" {
    command = <<EOF
cd lambda
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
  lambda_source_dir = "${path.module}/lambda"
  account_id       = var.account_id
  enable_dlq       = true

  tags = local.common_tags

  depends_on = [null_resource.build_lambda]
}

############################################
# S3 → Lambda Notification (Prefix Filter)
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
    filter_prefix       = "input/"
  }

  depends_on = [
    module.lambda_trigger,
    aws_lambda_permission.s3_invoke
  ]
}

############################################
# Optional EventBridge Schedule → Lambda
############################################

resource "aws_cloudwatch_event_rule" "glue_schedule" {
  count = var.enable_schedule ? 1 : 0

  name                = "${var.app_name}-${var.environment}-glue-schedule"
  schedule_expression = var.schedule_expression
  tags                = local.common_tags
}

resource "aws_cloudwatch_event_target" "glue_schedule_target" {
  count = var.enable_schedule ? 1 : 0

  rule      = aws_cloudwatch_event_rule.glue_schedule[0].name
  target_id = "lambda"
  arn       = module.lambda_trigger.lambda_arn
}

resource "aws_lambda_permission" "eventbridge_invoke" {
  count = var.enable_schedule ? 1 : 0

  statement_id  = "AllowEventBridgeInvoke"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_trigger.lambda_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.glue_schedule[0].arn
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

output "glue_job_name" {
  value = module.glue_job.glue_job_name
}

output "lambda_trigger_name" {
  value = module.lambda_trigger.lambda_name
}
