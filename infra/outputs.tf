############################################
# ALB / ECS / Glue / Lambda Outputs
############################################

output "alb_dns_name" {
  value = module.alb.alb_dns_name
}

output "ecs_service_id" {
  value = module.ecs_service.ecs_service_id
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

############################################
# Compliance DB (RDS PostgreSQL)
############################################

output "compliance_db_endpoint" {
  value = module.compliance_db.db_endpoint
}

############################################
# DynamoDB Compliance Rules Table
############################################

output "compliance_dynamodb_table_name" {
  value = module.compliance_dynamodb.table_name
}

output "compliance_dynamodb_table_arn" {
  value = module.compliance_dynamodb.table_arn
}

############################################
# Redis Compliance Cache (Multi-AZ)
############################################

output "compliance_redis_primary_endpoint" {
  value = module.compliance_redis.primary_endpoint
}

output "compliance_redis_reader_endpoint" {
  value = module.compliance_redis.reader_endpoint
}

output "compliance_redis_security_group_id" {
  value = module.compliance_redis.security_group_id
}
