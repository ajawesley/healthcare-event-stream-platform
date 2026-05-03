output "bucket_name" {
  description = "Name of the HESP raw event S3 bucket."
  value       = aws_s3_bucket.this.bucket
}

output "bucket_arn" {
  description = "ARN of the HESP raw event S3 bucket. Pass to the IAM module to scope the ingest task role policy."
  value       = aws_s3_bucket.this.arn
}

output "kms_key_arn" {
  description = "ARN of the KMS CMK used to encrypt raw event objects. Pass to the IAM module to grant kms:GenerateDataKey and kms:Decrypt to the ingest task role."
  value       = aws_kms_key.this.arn
}

output "kms_key_id" {
  description = "Key ID of the KMS CMK. Use for key policy references and CloudTrail event filtering."
  value       = aws_kms_key.this.key_id
}

output "kms_alias_arn" {
  description = "ARN of the KMS alias for the raw event bucket CMK. Use for human-readable key references in IAM policies."
  value       = aws_kms_alias.this.arn
}
