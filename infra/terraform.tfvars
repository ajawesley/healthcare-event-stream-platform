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

# IAM module inputs (CORRECT)
raw_bucket_arn    = "arn:aws:s3:::hesp-dev-raw-events-001"
script_bucket_arn = "arn:aws:s3:::hesp-dev-glue-scripts-001"
golden_bucket_arn = "arn:aws:s3:::hesp-dev-golden-events-001"

# MUST replace with real CMK ARN
kms_key_arn       = "arn:aws:kms:us-east-1:045797643729:key/arn:aws:kms:us-east-1:045797643729:key/fb903a76-6d90-4ae0-a17b-1f2fc64faa4b"
