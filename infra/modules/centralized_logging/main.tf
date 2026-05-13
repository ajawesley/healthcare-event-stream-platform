############################################
# CloudWatch Logs → S3 Centralized Logging
############################################

resource "aws_iam_role" "cw_to_s3_role" {
  name = "${var.name_prefix}-cw-to-s3-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })

  tags = var.tags
}

resource "aws_iam_role_policy" "cw_to_s3_policy" {
  name = "${var.name_prefix}-cw-to-s3-policy"
  role = aws_iam_role.cw_to_s3_role.id

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
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:PutObjectAcl"
        ]
        Resource = "${var.log_archive_bucket_arn}/*"
      }
    ]
  })
}

############################################
# Lambda Forwarder (CloudWatch → S3)
############################################

resource "aws_lambda_function" "cw_forwarder" {
  function_name = "${var.name_prefix}-cw-forwarder"
  role          = aws_iam_role.cw_to_s3_role.arn
  handler       = "index.handler"
  runtime       = "python3.11"

  filename         = var.lambda_zip_path
  source_code_hash = filebase64sha256(var.lambda_zip_path)

  timeout = 30

  tags = var.tags
}

############################################
# CloudWatch Log Subscription Filters
############################################

resource "aws_cloudwatch_log_subscription_filter" "forward_all" {
  count = length(var.log_group_names)

  name            = "${var.name_prefix}-subscription-${count.index}"
  log_group_name  = var.log_group_names[count.index]
  filter_pattern  = ""
  destination_arn = aws_lambda_function.cw_forwarder.arn
}
