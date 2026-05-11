module "compliance_dynamodb" {
  source = "./modules/dynamodb_compliance_rules"

  table_name = "${var.app_name}-${var.environment}-compliance-rules"

  ttl_enabled        = false
  ttl_attribute_name = "expires_at"

  tags = local.base_tags
}
