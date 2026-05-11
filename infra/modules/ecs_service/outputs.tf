############################################
# ECS Service Outputs
############################################

output "ecs_service_name" {
  description = "Name of the ECS service"
  value       = aws_ecs_service.this.name
}

output "ecs_service_id" {
  description = "ID of the ECS service (cluster/service-name)"
  value       = aws_ecs_service.this.id
}

output "ecs_service_cluster_arn" {
  description = "ARN of the ECS cluster the service runs in"
  value       = aws_ecs_service.this.cluster
}

output "ecs_task_definition_arn" {
  description = "ARN of the ECS task definition"
  value       = aws_ecs_task_definition.this.arn
}

output "ecs_service_sg_ids" {
  description = "Security groups actually attached to the ECS service"
  value       = aws_ecs_service.this.network_configuration[0].security_groups
}
