############################################
# S3 Bucket Module Outputs
############################################

output "raw_bucket_name" {
  description = "Raw events bucket name"
  value       = aws_s3_bucket.raw.bucket
}

output "raw_bucket_arn" {
  description = "Raw events bucket ARN"
  value       = aws_s3_bucket.raw.arn
}

output "golden_bucket_name" {
  description = "Golden events bucket name"
  value       = aws_s3_bucket.golden.bucket
}

output "golden_bucket_arn" {
  description = "Golden events bucket ARN"
  value       = aws_s3_bucket.golden.arn
}

output "scripts_bucket_name" {
  description = "Glue scripts bucket name"
  value       = aws_s3_bucket.scripts.bucket
}

output "scripts_bucket_arn" {
  description = "Glue scripts bucket ARN"
  value       = aws_s3_bucket.scripts.arn
}

output "access_logs_bucket_name" {
  description = "Access logs bucket name"
  value       = aws_s3_bucket.access_logs.bucket
}

output "access_logs_bucket_arn" {
  description = "Access logs bucket ARN"
  value       = aws_s3_bucket.access_logs.arn
}

output "log_archive_bucket_name" {
  description = "CloudTrail/Config/GuardDuty log archive bucket name"
  value       = aws_s3_bucket.log_archive.bucket
}

output "log_archive_bucket_arn" {
  description = "CloudTrail/Config/GuardDuty log archive bucket ARN"
  value       = aws_s3_bucket.log_archive.arn
}
