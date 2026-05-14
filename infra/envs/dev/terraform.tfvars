app_name    = "hesp"
environment = "dev"
aws_region  = "us-east-1"

owner       = "ajamu"
cost_center = "12345"

raw_bucket_name         = "hesp-dev-raw-events-001"
access_logs_bucket_name = "hesp-dev-access-logs-001"

container_image = "045797643729.dkr.ecr.us-east-1.amazonaws.com/hesp:latest"

desired_count = 1

compliance_db_password_secret_arn = "arn:aws:secretsmanager:us-east-1:045797643729:secret:hesp/compliance-db-password-vOL4A7"
compliance_db_name                = "hesp_compliance"
compliance_db_username            = "ajawe"

s3_output_base_path = "s3://hesp-dev-golden-events-001/golden-events/"
s3_error_path       = "s3://hesp-dev-golden-events-001/errors/"

glue_script_s3_path = "scripts/glue_job.py"
glue_temp_dir       = "tmp/"

lambda_zip_path = "../../../cmd/lambda/lambda.zip"

log_archive_bucket_name = "acmecorp-log-archive-001"


github_oidc_role_arn = "arn:aws:iam::045797643729:role/acme-github-oidc-deploy-role"

workload_account_id = "111111111111"

adot_image = "045797643729.dkr.ecr.us-east-1.amazonaws.com/hesp-adot:latest"
