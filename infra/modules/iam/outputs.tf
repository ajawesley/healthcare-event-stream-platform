output "execution_role_arn" {
  value = aws_iam_role.ecs_execution.arn
}

output "task_role_arn" {
  value = aws_iam_role.ecs_task.arn
}

output "glue_role_arn" {
  value = aws_iam_role.glue.arn
}

output "glue_role_name" {
  value = aws_iam_role.glue.name
}

output "lambda_role_arn" {
  value = aws_iam_role.lambda.arn
}

output "lambda_role_name" {
  value = aws_iam_role.lambda.name
}

output "cloudtrail_s3_role_arn" {
  value = aws_iam_role.cloudtrail_s3_role.arn
}

output "config_role_arn" {
  description = "IAM role ARN for AWS Config Recorder"
  value       = aws_iam_role.config.arn
}
