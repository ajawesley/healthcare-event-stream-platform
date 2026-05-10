############################################
# Glue Job Module
############################################

module "glue_job" {
  source = "./modules/glue_job"

  app_name    = var.app_name
  environment = var.environment
  owner       = var.owner
  cost_center = var.cost_center

  glue_role_arn = module.iam.glue_role_arn

  script_bucket  = var.script_bucket
  script_s3_path = var.glue_script_s3_path
  temp_dir       = var.glue_temp_dir

  raw_bucket    = aws_s3_bucket.this.bucket
  golden_bucket = "${var.app_name}-${var.environment}-golden-events-001"

  log_group_name = "/aws/glue/${var.app_name}-${var.environment}"
  tags           = local.base_tags
}

############################################
# Glue Crawlers Module
############################################

module "glue_crawlers" {
  source      = "./modules/glue_crawlers"
  environment = var.environment
  tags        = local.base_tags

  events_bucket = "${var.app_name}-${var.environment}-golden-events-001"
  errors_bucket = "${var.app_name}-${var.environment}-golden-events-001/errors"
}
