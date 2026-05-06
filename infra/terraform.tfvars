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

# ECS service configuration
container_image = "045797643729.dkr.ecr.us-east-1.amazonaws.com/hesp:latest"
desired_count   = 1

# Glue job script + temp directory (CORRECTED)
script_bucket       = "hesp-dev-glue-scripts-001"
glue_script_s3_path = "s3://hesp-dev-glue-scripts-001/scripts/glue_job.py"
glue_temp_dir       = "s3://hesp-dev-glue-scripts-001/tmp/"

