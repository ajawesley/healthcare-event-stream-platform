locals {
  base_tags = merge(
    var.tags,
    {
      App         = var.app_name
      Environment = var.environment
      Owner       = var.owner
      CostCenter  = var.cost_center
      ManagedBy   = "terraform"
    }
  )
}

############################################
# Lambda Function (Custom Runtime)
############################################

resource "aws_lambda_function" "this" {
  function_name = "${var.app_name}-${var.environment}-trigger"
  role          = var.lambda_role_arn

  # Custom runtime uses "bootstrap" as the handler
  handler = "bootstrap"
  runtime = "provided.al2023"

  architectures    = ["arm64"]  # only support arm64

  filename         = var.lambda_zip_path
  source_code_hash = filebase64sha256(var.lambda_zip_path)

  environment {
    variables = {
      GLUE_JOB_NAME = var.glue_job_name
      RAW_BUCKET    = var.raw_bucket_name
    }
  }

  tags = local.base_tags
}

############################################
# IAM Policy: Allow Lambda to Start Glue Job
############################################

resource "aws_iam_policy" "lambda_glue_policy" {
  name = "${var.app_name}-${var.environment}-lambda-glue"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
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
# IAM Policy: Allow Lambda to Write Logs
############################################

resource "aws_iam_policy" "lambda_logging_policy" {
  name = "${var.app_name}-${var.environment}-lambda-logging"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:*:*:*"
      }
    ]
  })
}

############################################
# IAM Policy: Allow Lambda to Read S3
############################################
resource "aws_iam_policy" "lambda_s3_read" {
  name = "${var.app_name}-${var.environment}-lambda-s3-read"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws:s3:::${var.raw_bucket_name}",
          "arn:aws:s3:::${var.raw_bucket_name}/*"
        ]
      }
    ]
  })
}

############################################
# Attach Policies to Lambda Role
############################################

resource "aws_iam_role_policy_attachment" "lambda_glue_attach" {
  role       = var.lambda_role_name   # FIXED: no brittle ARN splitting
  policy_arn = aws_iam_policy.lambda_glue_policy.arn

  depends_on = [aws_iam_policy.lambda_glue_policy]
}

resource "aws_iam_role_policy_attachment" "lambda_logging_attach" {
  role       = var.lambda_role_name
  policy_arn = aws_iam_policy.lambda_logging_policy.arn

  depends_on = [aws_iam_policy.lambda_logging_policy]
}

resource "aws_iam_role_policy_attachment" "lambda_s3_read_attach" {
  role       = var.lambda_role_name
  policy_arn = aws_iam_policy.lambda_s3_read.arn
}
