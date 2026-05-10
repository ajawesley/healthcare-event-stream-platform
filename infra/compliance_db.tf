############################################
# Compliance Database (RDS PostgreSQL)
############################################

module "compliance_db" {
  source = "./modules/rds_postgres_compliance_db"

  name                    = "${var.app_name}-${var.environment}-compliance-db"
  vpc_id                  = module.vpc.vpc_id
  isolated_subnet_ids     = module.vpc.isolated_subnets
  ingestion_service_sg_id = aws_security_group.ecs.id

  db_username = var.compliance_db_username
  db_password = var.compliance_db_password

  tags = local.base_tags
}
