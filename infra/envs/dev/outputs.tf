############################################
# ALB
############################################

output "alb_dns" {
  description = "DNS name of the ALB"
  value       = module.alb.alb_dns
}

############################################
# ECS
############################################

output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = aws_ecs_cluster.cluster.name
}

############################################
# S3 Buckets
############################################

output "raw_bucket_name" {
  description = "Raw events bucket name"
  value       = module.s3_buckets.raw_bucket_name
}

output "golden_bucket_name" {
  description = "Golden events bucket name"
  value       = module.s3_buckets.golden_bucket_name
}

############################################
# Debugging / Validation
############################################

output "github_oidc_role_arn" {
  description = "Org-level GitHub OIDC deploy role ARN used by this environment"
  value       = var.github_oidc_role_arn
}
