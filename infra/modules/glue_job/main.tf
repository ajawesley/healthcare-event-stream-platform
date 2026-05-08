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
    "--extra-py-files"                   = "s3://${var.script_bucket}/scripts/lib/job_lib.zip"

    # STATIC ONLY — Lambda overrides input_path at runtime
    "--output_base_path" = "s3://${var.golden_bucket}/"
    "--error_path"       = "s3://${var.golden_bucket}/errors/"

    # Lineage-aware arguments (Glue job reads these from S3 JSON)
    "--raw_bucket"       = var.raw_bucket
    "--golden_bucket"    = var.golden_bucket
  }

  glue_version      = "4.0"
  number_of_workers = 2
  worker_type       = "G.1X"
  timeout           = 60
  max_retries       = 1

  tags = local.base_tags

  depends_on = [
    aws_cloudwatch_log_group.glue
  ]
}
