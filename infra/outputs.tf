output "alb_dns_name" {
  value = module.alb.alb_dns_name
}

output "ecs_service_id" {
  value = module.ecs_service.service_id
}

output "s3_bucket_name" {
  value = module.s3.bucket_name
}

output "glue_job_name" {
  value = module.glue_job.glue_job_name
}

output "glue_job_arn" {
  value = module.glue_job.glue_job_arn
}

output "lambda_trigger_name" {
  value = module.lambda_trigger.lambda_name
}

output "lambda_trigger_arn" {
  value = module.lambda_trigger.lambda_arn
}

output "github_oidc_role_arn" {
  value = module.github_oidc.role_arn
}

