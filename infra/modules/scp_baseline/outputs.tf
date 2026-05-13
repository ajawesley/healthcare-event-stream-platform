output "scp_names" {
  description = "Names of all SCPs created"
  value       = [for p in aws_organizations_policy.scp : p.name]
}
