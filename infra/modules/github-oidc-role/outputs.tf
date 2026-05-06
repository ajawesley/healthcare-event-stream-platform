output "role_arn" {
  description = "ARN of the IAM role GitHub Actions will assume."
  value       = aws_iam_role.this.arn
}

output "oidc_provider_arn" {
  description = "ARN of the GitHub OIDC provider."
  value       = aws_iam_openid_connect_provider.github.arn
}
