output "aggregator_id" {
  description = "ID of the Config aggregator"
  value       = aws_config_configuration_aggregator.org.id
}
