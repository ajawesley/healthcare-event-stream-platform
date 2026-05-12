include {
  path = find_in_parent_folders()
}

terraform {
  source = "../workloads"
}

inputs = {
  aws_region              = "us-east-1"
  org_name                = "hesp"
  workloads_account_id    = "WORKLOADS_ACCOUNT_ID"
  log_archive_bucket_name = "hesp-log-archive"
  log_archive_bucket_arn  = "arn:aws:s3:::hesp-log-archive"
  config_role_arn         = "arn:aws:iam::WORKLOADS_ACCOUNT_ID:role/ConfigRecorderRole"
  security_contact_email  = "security@hesp.com"
  account_alias           = "workloads"
  owner                   = "ajamu"
}
