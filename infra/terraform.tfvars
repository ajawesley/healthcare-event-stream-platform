aws_region = "us-east-1"

app_name    = "hesp"
environment = "dev"
owner       = "platform-team"
cost_center = "12345"

bucket_name            = "hesp-dev-raw-events-001"
access_log_bucket_name = "hesp-dev-access-logs"

container_image = "045797643729.dkr.ecr.us-east-1.amazonaws.com/hesp-dev-ingest:latest"

desired_count = 1

script_bucket       = "hesp-dev-glue-scripts-001"
glue_script_s3_path = "s3://hesp-dev-glue-scripts-001/scripts/glue_job.py"
glue_temp_dir       = "s3://hesp-dev-glue-scripts-001/tmp/"

# --- Observability Vendor Keys ---
dd_api_key        = "" # optional
honeycomb_api_key = "" # optional
honeycomb_dataset = "" # optional

# --- NEW: ACM Certificate ARN (Option A workflow) ---
acm_certificate_arn = "REPLACE_ME_WITH_REAL_ACM_ARN"
