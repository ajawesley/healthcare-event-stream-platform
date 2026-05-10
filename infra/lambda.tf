############################################
# Lambda Build
############################################

resource "null_resource" "build_lambda" {
  triggers = {
    src_hash = filemd5("${path.root}/../cmd/lambda/main.go")
  }

  provisioner "local-exec" {
    command = <<EOF
cd ../cmd/lambda
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap main.go
zip lambda.zip bootstrap
EOF
  }
}

############################################
# Lambda Trigger Module
############################################

module "lambda_trigger" {
  source = "./modules/lambda_trigger"

  app_name    = var.app_name
  environment = var.environment
  owner       = var.owner
  cost_center = var.cost_center

  output_base_path = var.s3_output_base_path
  error_path       = var.s3_error_path

  glue_job_name = module.glue_job.glue_job_name
  glue_job_arn  = module.glue_job.glue_job_arn

  raw_bucket_name = aws_s3_bucket.this.bucket

  lambda_role_arn  = module.iam.lambda_role_arn
  lambda_role_name = module.iam.lambda_role_name

  lambda_zip_path = "${path.root}/../cmd/lambda/lambda.zip"

  kms_key_arn = aws_kms_key.this.arn
  tags        = local.base_tags

  depends_on = [null_resource.build_lambda]
}

############################################
# S3 → Lambda Notification
############################################

resource "aws_lambda_permission" "s3_invoke" {
  statement_id  = "AllowS3Invoke"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_trigger.lambda_name
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.this.arn
}

resource "aws_s3_bucket_notification" "raw_events_trigger" {
  bucket = aws_s3_bucket.this.bucket

  lambda_function {
    lambda_function_arn = module.lambda_trigger.lambda_arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "events/"
  }

  depends_on = [
    module.lambda_trigger,
    aws_lambda_permission.s3_invoke
  ]
}
