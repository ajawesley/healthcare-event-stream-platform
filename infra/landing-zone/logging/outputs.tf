############################################
# Log Archive Bucket
############################################

output "log_archive_bucket_name" {
  description = "Name of the centralized log archive bucket"
  value       = aws_s3_bucket.log_archive.bucket
}

output "log_archive_bucket_arn" {
  description = "ARN of the centralized log archive bucket"
  value       = aws_s3_bucket.log_archive.arn
}

############################################
# KMS Key for Logging
############################################

output "log_kms_key_arn" {
  description = "ARN of the KMS key used for log encryption"
  value       = aws_kms_key.logs.arn
}

############################################
# CloudTrail
############################################

output "org_cloudtrail_name" {
  description = "Name of the organization-level CloudTrail"
  value       = aws_cloudtrail.org_trail.name
}

############################################
# AWS Config Aggregator
############################################

output "org_config_aggregator_name" {
  description = "Name of the organization-level AWS Config aggregator"
  value       = aws_config_configuration_aggregator.org.name
}
