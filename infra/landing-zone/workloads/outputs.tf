output "vpc_id" {
  value = module.vpc_baseline.vpc_id
}

output "public_subnets" {
  value = module.vpc_baseline.public_subnet_ids
}

output "private_subnets" {
  value = module.vpc_baseline.private_subnet_ids
}

output "guardduty_detector_id" {
  value = aws_guardduty_detector.this.id
}

output "config_recorder_name" {
  value = aws_config_configuration_recorder.this.name
}
