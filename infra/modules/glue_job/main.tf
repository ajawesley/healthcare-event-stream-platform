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
# CloudWatch Log Group for Glue Job
############################################

resource "aws_cloudwatch_log_group" "glue" {
  name              = var.log_group_name
  retention_in_days = 30
  tags              = local.base_tags
}

############################################
# (Optional) Glue Security Configuration
############################################
# resource "aws_glue_security_configuration" "this" {
#   name = "${var.app_name}-${var.environment}-glue-sec"
#
#   encryption_configuration {
#     cloudwatch_encryption {
#       cloudwatch_encryption_mode = "SSE-KMS"
#       kms_key_arn                = var.kms_key_arn
#     }
#     job_bookmarks_encryption {
#       job_bookmarks_encryption_mode = "CSE-KMS"
#       kms_key_arn                   = var.kms_key_arn
#     }
#     s3_encryption {
#       s3_encryption_mode = "SSE-KMS"
#       kms_key_arn        = var.kms_key_arn
#     }
#   }
# }

############################################
# Glue Job
############################################

resource "aws_glue_job" "this" {
  name     = "${var.app_name}-${var.environment}-job"
  role_arn = var.glue_role_arn

  command {
    name            = "glueetl"
    script_location = var.script_s3_path
    python_version  = "3"
  }

  default_arguments = {
    "--TempDir"                          = var.temp_dir
    "--enable-continuous-cloudwatch-log" = "true"
    "--continuous-log-logGroup"          = var.log_group_name
    "--job-language"                     = "python"
  }

  glue_version      = "4.0"
  number_of_workers = 2
  worker_type       = "G.1X"
  timeout           = 60
  max_retries       = 1

  # security_configuration = aws_glue_security_configuration.this.name

  tags = local.base_tags

  depends_on = [
    aws_cloudwatch_log_group.glue
  ]
}
