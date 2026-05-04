# AWS region for all resources
aws_region = "us-east-1"

# Application metadata
app_name    = "hesp"
environment = "dev"
owner       = "platform-team"
cost_center = "12345"

# S3 bucket names (must be globally unique when applied)
bucket_name            = "hesp-dev-raw-events-001"
access_log_bucket_name = "hesp-dev-access-logs"

# Networking
vpc_id = "vpc-053e0aaa362d0b71e"

public_subnet_ids = [
  "subnet-01f36b9ed440b5613",
  "subnet-065421a8ee06a681b"
]

private_subnet_ids = [
  "subnet-08b6a642d7bd77695",
  "subnet-025fc8bf74d1fd4ef"
]

# ECS service configuration
container_image = "public.ecr.aws/xxxxxxx/ingest-service:latest"
desired_count   = 1

# Glue job IAM role (must exist)
glue_role_arn = "arn:aws:iam::123456789012:role/hesp-dev-glue-role"

# Optional EventBridge schedule (disabled by default)
enable_schedule     = false
schedule_expression = "rate(1 day)"

# Tags
tags = {
  "environment"         = "dev"
  "owner"               = "platform-team"
  "managed-by"          = "terraform"
  "cost-center"         = "12345"
  "data-classification" = "phi"
}
