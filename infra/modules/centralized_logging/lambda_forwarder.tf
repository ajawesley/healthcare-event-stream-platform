############################################
# Lambda Permissions for CloudWatch Logs
############################################

resource "aws_lambda_permission" "allow_cw" {
  count = length(var.log_group_names)

  statement_id  = "AllowExecutionFromCW-${count.index}"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.cw_forwarder.function_name
  principal     = "logs.amazonaws.com"

  source_arn = "arn:aws:logs:${var.aws_region}:${var.account_id}:log-group:${var.log_group_names[count.index]}:*"
}
