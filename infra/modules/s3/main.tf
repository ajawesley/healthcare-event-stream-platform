locals {
  # Required tags are merged last so callers cannot accidentally override them.
  required_tags = {
    environment           = var.environment
    "data-classification" = "phi"
    owner                 = var.owner
    "cost-center"         = var.cost_center
    "managed-by"          = "terraform"
  }
  tags = merge(var.tags, local.required_tags)

  # Object Lock mode is COMPLIANCE in prod, GOVERNANCE in dev/staging.
  # COMPLIANCE: immutable even for root. GOVERNANCE: admin override possible.
  object_lock_mode = var.environment == "prod" ? "COMPLIANCE" : "GOVERNANCE"
}

# -----------------------------------------------------------------------------
# KMS Customer Managed Key
# Provides key-use audit trail via CloudTrail, annual rotation, and the ability
# to disable the key to render archived PHI inaccessible under a legal hold.
# -----------------------------------------------------------------------------
resource "aws_kms_key" "this" {
  description             = "HESP ${var.environment} raw event bucket CMK"
  deletion_window_in_days = 30
  enable_key_rotation     = true
  tags                    = local.tags
}

resource "aws_kms_alias" "this" {
  name          = "alias/hesp/${var.environment}/raw-events"
  target_key_id = aws_kms_key.this.key_id
}

# -----------------------------------------------------------------------------
# S3 Bucket
# object_lock_enabled must be set at creation time — it cannot be added later.
# prevent_destroy guards against accidental destruction of PHI data.
# -----------------------------------------------------------------------------
resource "aws_s3_bucket" "this" {
  bucket              = var.bucket_name
  object_lock_enabled = true

  tags = local.tags

  lifecycle {
    prevent_destroy = true
  }
}

# -----------------------------------------------------------------------------
# Public Access Block
# All four flags required. Prevents any future bucket ACL or policy change
# from inadvertently exposing PHI objects publicly.
# -----------------------------------------------------------------------------
resource "aws_s3_bucket_public_access_block" "this" {
  bucket = aws_s3_bucket.this.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# -----------------------------------------------------------------------------
# Versioning
# Required prerequisite for Object Lock. Enables point-in-time recovery.
# -----------------------------------------------------------------------------
resource "aws_s3_bucket_versioning" "this" {
  bucket = aws_s3_bucket.this.id

  versioning_configuration {
    status = "Enabled"
  }
}

# -----------------------------------------------------------------------------
# Server-Side Encryption — SSE-KMS with CMK
# bucket_key_enabled batches KMS API calls per-bucket rather than per-object,
# reducing KMS request cost significantly at ingest volume.
# -----------------------------------------------------------------------------
resource "aws_s3_bucket_server_side_encryption_configuration" "this" {
  bucket = aws_s3_bucket.this.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.this.arn
    }
    bucket_key_enabled = true
  }
}

# -----------------------------------------------------------------------------
# Object Lock — default retention policy
# COMPLIANCE (prod): no principal including root can delete before 7 years.
# GOVERNANCE (dev/staging): principals with s3:BypassGovernanceRetention can
# override, allowing test data cleanup without a full account escalation.
# -----------------------------------------------------------------------------
resource "aws_s3_bucket_object_lock_configuration" "this" {
  bucket = aws_s3_bucket.this.id

  rule {
    default_retention {
      mode  = local.object_lock_mode
      years = 7
    }
  }
}

# -----------------------------------------------------------------------------
# Lifecycle Configuration
# Transitions objects to Intelligent-Tiering after 30 days to reduce storage
# cost for infrequently accessed historical PHI events.
# -----------------------------------------------------------------------------
resource "aws_s3_bucket_lifecycle_configuration" "this" {
  bucket = aws_s3_bucket.this.id

  rule {
    id     = "intelligent-tiering"
    status = "Enabled"

    filter {}

    transition {
      days          = 30
      storage_class = "INTELLIGENT_TIERING"
    }
  }
}

# -----------------------------------------------------------------------------
# Server Access Logging
# Captures every GET, PUT, DELETE, and HEAD request with requester identity,
# timestamp, and object key. Required for HIPAA §164.312(b) audit controls.
# Logs are written to the centralised audit bucket managed by the root module.
# -----------------------------------------------------------------------------
resource "aws_s3_bucket_logging" "this" {
  bucket        = aws_s3_bucket.this.id
  target_bucket = var.access_log_bucket_id
  target_prefix = "s3-access-logs/${var.bucket_name}/"
}

# -----------------------------------------------------------------------------
# Bucket Policy
# Three statements, applied in order:
#   1. DenyNonTLS         — rejects all requests over plain HTTP.
#   2. DenyUnencryptedPut — rejects any PutObject not using SSE-KMS, ensuring
#                           objects cannot be stored without encryption even if
#                           the SSE configuration above were somehow bypassed.
#   3. AllowIngestPutOnly — explicit allow for the ECS task role; PutObject
#                           only, no read or delete rights.
# -----------------------------------------------------------------------------
resource "aws_s3_bucket_policy" "this" {
  bucket = aws_s3_bucket.this.id

  # Ensure public access block is in place before the policy is applied.
  depends_on = [aws_s3_bucket_public_access_block.this]

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "DenyNonTLS"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:*"
        Resource = [
          aws_s3_bucket.this.arn,
          "${aws_s3_bucket.this.arn}/*"
        ]
        Condition = {
          Bool = {
            "aws:SecureTransport" = "false"
          }
        }
      },
      {
        Sid       = "DenyUnencryptedPut"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:PutObject"
        Resource  = "${aws_s3_bucket.this.arn}/*"
        Condition = {
          StringNotEquals = {
            "s3:x-amz-server-side-encryption" = "aws:kms"
          }
        }
      },
      {
        Sid    = "AllowIngestPutOnly"
        Effect = "Allow"
        Principal = {
          AWS = var.ingest_task_role_arn
        }
        Action   = "s3:PutObject"
        Resource = "${aws_s3_bucket.this.arn}/*"
      }
    ]
  })
}
