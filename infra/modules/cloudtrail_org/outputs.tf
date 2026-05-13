output "cloudtrail_arn" {
  description = "ARN of the organization-wide CloudTrail"
  value       = aws_cloudtrail.org.arn
}
