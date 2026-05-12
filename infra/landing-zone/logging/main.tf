terraform {
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
# Locals
############################################

locals {
  log_bucket_name = var.log_archive_bucket_name
  tags = merge(
    {
      Project = var.org_name
      Owner   = var.owner
    },
    var.extra_tags
  )
}

############################################
# KMS Key for Log Archive
############################################

resource "aws_kms_key" "logs" {
  description             = "KMS key for org-level logging"
  deletion_window_in_days = 30
  enable_key_rotation     = true

  tags = local.tags
}

resource "aws_kms_alias" "logs" {
  name          = "alias/${var.org_name}-log-archive-kms"
  target_key_id = aws_kms_key.logs.key_id
}

############################################
# Log Archive Bucket
############################################

resource "aws_s3_bucket" "log_archive" {
  bucket = local.log_bucket_name

  lifecycle {
    prevent_destroy = true
  }

  tags = local.tags
}

resource "aws_s3_bucket_versioning" "log_archive" {
  bucket = aws_s3_bucket.log_archive.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "log_archive" {
  bucket = aws_s3_bucket.log_archive.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.logs.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "log_archive" {
  bucket = aws_s3_bucket.log_archive.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

############################################
# Bucket Policy (CloudTrail + Config)
############################################

resource "aws_s3_bucket_policy" "log_archive" {
  bucket = aws_s3_bucket.log_archive.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      # Allow CloudTrail to write logs
      {
        Sid    = "AWSCloudTrailWrite"
        Effect = "Allow"
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        }
        Action   = "s3:PutObject"
        Resource = "${aws_s3_bucket.log_archive.arn}/AWSLogs/*"
        Condition = {
          StringEquals = {
            "s3:x-amz-acl" = "bucket-owner-full-control"
          }
        }
      },

      # Allow CloudTrail to read bucket ACL
      {
        Sid    = "AWSCloudTrailAclCheck"
        Effect = "Allow"
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        }
        Action   = "s3:GetBucketAcl"
        Resource = aws_s3_bucket.log_archive.arn
      },

      # Allow AWS Config to write snapshots
      {
        Sid    = "AWSConfigWrite"
        Effect = "Allow"
        Principal = {
          Service = "config.amazonaws.com"
        }
        Action   = "s3:PutObject"
        Resource = "${aws_s3_bucket.log_archive.arn}/AWSLogs/*"
      }
    ]
  })
}

############################################
# Org-Level CloudTrail
############################################

resource "aws_cloudtrail" "org_trail" {
  name                          = "${var.org_name}-org-trail"
  s3_bucket_name                = aws_s3_bucket.log_archive.id
  is_organization_trail         = true
  include_global_service_events = true
  is_multi_region_trail         = true
  enable_log_file_validation    = true
  kms_key_id                    = aws_kms_key.logs.arn

  tags = local.tags
}

############################################
# Org-Level AWS Config Aggregator
############################################

resource "aws_config_configuration_aggregator" "org" {
  name = "${var.org_name}-org-aggregator"

  organization_aggregation_source {
    role_arn    = var.org_config_role_arn
    all_regions = true
  }

  depends_on = [
    aws_s3_bucket.log_archive,
    aws_kms_key.logs
  ]
}
