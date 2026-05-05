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
    "--TempDir"                  = var.temp_dir
    "--enable-continuous-cloudwatch-log" = "true"
    "--continuous-log-logGroup"          = var.log_group_name
  }

  glue_version       = "4.0"
  number_of_workers  = 2
  worker_type        = "G.1X"
  timeout            = 60

  tags = local.base_tags
}

