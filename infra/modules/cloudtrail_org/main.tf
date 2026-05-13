############################################
# Organization-wide CloudTrail
############################################

resource "aws_cloudtrail" "org" {
  name                          = "${var.name_prefix}-org-cloudtrail"
  s3_bucket_name                = var.log_archive_bucket_name
  kms_key_id                    = var.kms_key_arn
  is_multi_region_trail         = true
  enable_log_file_validation    = true
  include_global_service_events = true
  is_organization_trail         = var.is_organization_trail

  event_selector {
    read_write_type           = "All"
    include_management_events = true
  }

  tags = merge(
    var.tags,
    {
      Name = "${var.name_prefix}-org-cloudtrail"
    }
  )
}

############################################
# Allow CloudTrail to use the S3 bucket
############################################

resource "aws_s3_bucket_policy" "cloudtrail" {
  bucket = var.log_archive_bucket_name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AWSCloudTrailAclCheck"
        Effect    = "Allow"
        Principal = { Service = "cloudtrail.amazonaws.com" }
        Action    = "s3:GetBucketAcl"
        Resource  = "arn:aws:s3:::${var.log_archive_bucket_name}"
      },
      {
        Sid       = "AWSCloudTrailWrite"
        Effect    = "Allow"
        Principal = { Service = "cloudtrail.amazonaws.com" }
        Action    = "s3:PutObject"
        Resource  = "arn:aws:s3:::${var.log_archive_bucket_name}/AWSLogs/*"
        Condition = {
          StringEquals = {
            "s3:x-amz-acl" = "bucket-owner-full-control"
          }
        }
      }
    ]
  })
}
