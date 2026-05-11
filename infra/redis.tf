module "compliance_redis" {
  source = "./modules/redis_compliance"

  name                    = "${var.app_name}-${var.environment}-compliance-redis"
  vpc_id                  = module.vpc.vpc_id
  isolated_subnet_ids     = module.vpc.isolated_subnets
  ingestion_service_sg_id = aws_security_group.ecs.id

  node_type                  = "cache.t4g.small"
  replicas_per_node_group    = 1
  transit_encryption_enabled = true

  tags = local.base_tags
}
