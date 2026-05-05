output "execution_role_arn" {
  value = aws_iam_role.ecs_execution.arn
}

output "task_role_arn" {
  value = aws_iam_role.ecs_task.arn
}

output "glue_role_arn" {
  value = aws_iam_role.glue.arn
}

output "lambda_role_arn" {
  value = aws_iam_role.lambda.arn
}

