output "permission_boundary_policy_arn" {
  value = module.permission_boundaries.permission_boundary_arn
}

output "oidc_provider_arn" {
  value = module.oidc.provider_arn
}
