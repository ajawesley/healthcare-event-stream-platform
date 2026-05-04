############################################
# Lambda Trigger Module
############################################

locals {
  required_tags = {
    App         = var.app_name
    Environment = var.environment
    Owner       = var.owner
    CostCenter  = var.cost_center
    ManagedBy   = "terraform"
  }

  tags = merge(var.tags, local.required_tags)
}

############################################
# IAM Role for Lambda
############################################

resource "aws_iam_role" "lambda" {
  name = "${var.app_name}-${var.environment}-lambda-trigger"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
      Action = "sts:AssumeRole"
      Condition = {
        StringEquals = {
          "aws:SourceAccount" = var.account_id
        }
      }
    }]
  })

  tags = local.tags
}

resource "aws_iam_role_policy" "lambda" {
  name = "${var.app_name}-${var.environment}-lambda-trigger-policy"
  role = aws_iam_role.lambda.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:*:*:log-group:/aws/lambda/${var.app_name}-${var.environment}-glue-trigger:*"
      },
      {
        Effect = "Allow"
        Action = [
          "glue:StartJobRun"
        ]
        Resource = var.glue_job_arn
      }
    ]
  })
}

############################################
# Package Lambda (fixes plan-time hash issue)
############################################

data "archive_file" "lambda_zip" {
  type        = "zip"
  source_dir  = var.lambda_source_dir
  output_path = "${path.module}/lambda.zip"
}

############################################
# Lambda Function
############################################

resource "aws_lambda_function" "this" {
  function_name = "${var.app_name}-${var.environment}-glue-trigger"
  role          = aws_iam_role.lambda.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["arm64"]

  filename         = data.archive_file.lambda_zip.output_path
  source_code_hash = data.archive_file.lambda_zip.output_base64sha256

  timeout     = 30
  memory_size = 128

  tracing_config {
    mode = "Active"
  }

  environment {
    variables = {
      GLUE_JOB_NAME = var.glue_job_name
      RAW_BUCKET    = var.raw_bucket_name
    }
  }

  tags = local.tags
}

############################################
# Optional DLQ (recommended for PHI)
############################################

resource "aws_sqs_queue" "dlq" {
  count = var.enable_dlq ? 1 : 0

  name                      = "${var.app_name}-${var.environment}-lambda-dlq"
  message_retention_seconds = 1209600 # 14 days
}

resource "aws_lambda_event_invoke_config" "dlq" {
  count = var.enable_dlq ? 1 : 0

  function_name = aws_lambda_function.this.function_name

  destination_config {
    on_failure {
      destination = aws_sqs_queue.dlq[0].arn
    }
  }
}

############################################
# Outputs
############################################

output "lambda_arn" {
  value = aws_lambda_function.this.arn
}

output "lambda_name" {
  value = aws_lambda_function.this.function_name
}
