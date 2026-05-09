############################################
# CloudTrail for S3 Object-Level Events
############################################

resource "aws_cloudwatch_log_group" "s3_data_events" {
  name              = "/${var.app_name}/${var.environment}/s3-data-events"
  retention_in_days = 30
  tags = {
    App         = var.app_name
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

resource "aws_iam_role" "cloudtrail_s3_role" {
  name = "${var.app_name}-${var.environment}-cloudtrail-s3-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "cloudtrail.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "cloudtrail_s3_policy" {
  name = "${var.app_name}-${var.environment}-cloudtrail-s3-policy"
  role = aws_iam_role.cloudtrail_s3_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "${aws_cloudwatch_log_group.s3_data_events.arn}:*"
      }
    ]
  })
}

resource "aws_cloudtrail" "s3_data_events" {
  name                          = "${var.app_name}-${var.environment}-s3-data-events"
  s3_bucket_name                = aws_s3_bucket.access_logs.bucket
  include_global_service_events = false
  is_multi_region_trail         = false
  enable_logging                = true
  cloud_watch_logs_group_arn    = "${aws_cloudwatch_log_group.s3_data_events.arn}:*"
  cloud_watch_logs_role_arn     = aws_iam_role.cloudtrail_s3_role.arn

  event_selector {
    read_write_type           = "WriteOnly"
    include_management_events = false

    data_resource {
      type = "AWS::S3::Object"
      values = [
        "${aws_s3_bucket.this.arn}/",                                        # raw bucket
        "arn:aws:s3:::${var.app_name}-${var.environment}-golden-events-001/" # golden bucket
      ]
    }
  }

  depends_on = [
    aws_iam_role.cloudtrail_s3_role,
    aws_iam_role_policy.cloudtrail_s3_policy
  ]
}
