############################################
# AWS Config Organization Aggregator
############################################

resource "aws_config_configuration_aggregator" "org" {
  name = var.name

  organization_aggregation_source {
    all_regions = var.all_regions
    role_arn    = var.aggregation_role_arn
  }

  tags = var.tags
}
