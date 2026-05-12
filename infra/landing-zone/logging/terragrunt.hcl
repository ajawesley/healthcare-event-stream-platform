include {
  path = find_in_parent_folders()
}

terraform {
  source = "../logging"
}

inputs = {
  aws_region              = "us-east-1"
  org_name                = "hesp"
  log_archive_bucket_name = "hesp-log-archive"
  org_config_role_arn     = "arn:aws:iam::ORG_ACCOUNT_ID:role/OrgConfigRole"
  owner                   = "ajamu"
}
