############################################
# Glue Job Module
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
# Glue Job
############################################

resource "aws_glue_job" "this" {
  name         = "${var.app_name}-${var.environment}-ingest"
  role_arn     = var.glue_role_arn
  glue_version = "4.0"

  # Allow scaling per environment
  number_of_workers = var.number_of_workers
  worker_type       = var.worker_type

  command {
    name            = "glueetl"
    script_location = var.script_s3_path
    python_version  = "3"
  }

  default_arguments = {
    "--enable-continuous-cloudwatch-log" = "true"
    "--continuous-log-logGroup"          = var.log_group_name
    "--enable-metrics"                   = "true"
    "--job-language"                     = "python"
    "--TempDir"                          = var.temp_dir

    # Critical: load helper modules
    "--extra-py-files" = join(",", var.extra_py_files)
  }

  execution_property {
    max_concurrent_runs = 1
  }

  timeout = var.timeout_minutes

  tags = local.tags
}

############################################
# Outputs
############################################

output "glue_job_name" {
  value = aws_glue_job.this.name
}

output "glue_job_arn" {
  value = aws_glue_job.this.arn
}
