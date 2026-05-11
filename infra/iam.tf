############################################
# IAM Module
############################################

module "iam" {
  source = "./modules/iam"

  app_name    = var.app_name
  environment = var.environment
  owner       = var.owner
  cost_center = var.cost_center

  compliance_db_password_secret_arn = var.compliance_db_password_secret_arn

  dynamodb_table_arn = module.compliance_dynamodb.table_arn

  raw_bucket_arn    = aws_s3_bucket.this.arn
  script_bucket_arn = "arn:aws:s3:::${var.script_bucket}"
  golden_bucket_arn = "arn:aws:s3:::${var.app_name}-${var.environment}-golden-events-001"

  kms_key_arn   = aws_kms_key.this.arn
  log_group_arn = aws_cloudwatch_log_group.ecs.arn
  tags          = local.base_tags
}
