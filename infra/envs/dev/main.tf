module "compliance_db" {
  source = "../../modules/rds_postgres_compliance_db"

  name                    = "compliance-db-dev"
  vpc_id                  = module.vpc.vpc_id
  isolated_subnet_ids     = module.vpc.isolated_subnet_ids
  ingestion_service_sg_id = module.ingestion_service.sg_id

  db_username = "compliance_user"
  db_password = "SuperSecurePassword123!"
}
