output "task_role_arn" {
  description = "ARN of the ECS ingest task role."
  value       = aws_iam_role.task.arn
}

output "execution_role_arn" {
  description = "ARN of the ECS task execution role."
  value       = aws_iam_role.execution.arn
}

output "glue_role_arn" {
  description = "ARN of the Glue job role."
  value       = aws_iam_role.glue.arn
}
