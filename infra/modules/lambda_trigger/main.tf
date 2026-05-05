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
# Lambda Function
############################################

resource "aws_lambda_function" "this" {
  function_name = "${var.app_name}-${var.environment}-trigger"
  role          = var.lambda_role_arn
  handler       = "bootstrap"
  runtime       = "provided.al2"

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
# Allow Lambda to Start Glue Job
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

resource "aws_iam_role_policy_attachment" "lambda_glue_attach" {
  role       = split("/", var.lambda_role_arn)[1]
  policy_arn = aws_iam_policy.lambda_glue_policy.arn
}

