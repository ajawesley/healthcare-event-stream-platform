output "arn" {
  description = "ARN of the CloudTrail KMS key"
  value       = aws_kms_key.this.arn
}
