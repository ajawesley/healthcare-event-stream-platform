output "ecs_task_role_arn" {
  value = aws_iam_role.ecs_task_role.arn
}

output "lambda_exec_role_arn" {
  value = aws_iam_role.lambda_exec_role.arn
}

output "glue_role_arn" {
  value = aws_iam_role.glue_role.arn
}
