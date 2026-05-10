############################################
# Observability Module
############################################

module "observability" {
  source = "./modules/observability"

  environment             = var.environment
  aws_region              = var.aws_region
  app_name                = var.app_name
  tags                    = var.tags
  ecs_cluster_name        = aws_ecs_cluster.cluster.name
  ecs_service_name        = "${var.app_name}-${var.environment}-svc"
  alb_arn                 = module.alb.alb_arn
  target_group_arn        = module.alb.target_group_arn
  raw_bucket_arn          = aws_s3_bucket.this.arn
  access_logs_bucket_name = aws_s3_bucket.access_logs.bucket

  cloudtrail_s3_role_arn = module.iam.cloudtrail_s3_role_arn
}

