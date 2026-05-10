aws_region = "us-east-1"

app_name    = "hesp"
environment = "dev"
owner       = "platform-team"
cost_center = "12345"

bucket_name            = "hesp-dev-raw-events-001"
access_log_bucket_name = "hesp-dev-access-logs"

container_image = "045797643729.dkr.ecr.us-east-1.amazonaws.com/hesp-dev-ingest:latest"

desired_count = 1

s3_output_base_path = "s3://hesp-dev-golden-events-001/golden-events/"

s3_error_path = "s3://hesp-dev-golden-events-001/errors/"

script_bucket       = "hesp-dev-glue-scripts-001"
glue_script_s3_path = "s3://hesp-dev-glue-scripts-001/scripts/glue_job.py"
glue_temp_dir       = "s3://hesp-dev-glue-scripts-001/tmp/"



# --- NEW: ACM Certificate ARN ---
acm_certificate_arn = "REPLACE_ME_WITH_REAL_ACM_ARN"

compliance_db_password_secret_arn = "arn:aws:secretsmanager:us-east-1:045797643729:secret:hesp/compliance-db-password-vOL4A7"
compliance_db_name                = "hesp_compliance"
compliance_db_username            = "ajawe"
compliance_db_host                = "hesp-dev-compliance-db.ccp046i8yfzy.us-east-1.rds.amazonaws.com"
compliance_db_port                = 5432


