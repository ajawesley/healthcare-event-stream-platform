output "github_oidc_role_arn" {
  description = "Org-level GitHub OIDC deploy role ARN"
  value       = module.oidc.github_deploy_role_arn
}

output "oidc_provider_arn" {
  description = "OIDC provider ARN"
  value       = module.oidc.oidc_provider_arn
}
