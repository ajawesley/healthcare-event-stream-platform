output "alb_dns_name" {
  value = module.alb.alb_dns_name
}

output "ecs_service_id" {
  value = module.ecs_service.service_id
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

output "glue_crawlers_database" {
  value = module.glue_crawlers.database_name
}

output "glue_events_crawler" {
  value = module.glue_crawlers.events_crawler_name
}

output "glue_errors_crawler" {
  value = module.glue_crawlers.errors_crawler_name
}


