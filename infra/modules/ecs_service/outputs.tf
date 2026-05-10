output "ecs_service_name" {
  value = aws_ecs_service.this.name
}

output "ecs_task_definition_arn" {
  value = aws_ecs_task_definition.this.arn
}

output "ecs_service_sg_ids" {
  value = var.security_group_ids
}

output "service_id" {
  value = aws_ecs_service.this.id
}

