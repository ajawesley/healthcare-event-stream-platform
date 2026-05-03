output "task_role_arn" {
  description = "ARN of the ECS ingest task role."
  value       = aws_iam_role.task.arn
}

output "execution_role_arn" {
  description = "ARN of the ECS task execution role."
  value       = aws_iam_role.execution.arn
}
