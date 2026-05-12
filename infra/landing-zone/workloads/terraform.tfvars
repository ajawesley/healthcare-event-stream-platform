aws_region              = "us-east-1"
org_name                = "hesp"
workloads_account_id    = "123456789012"
log_archive_bucket_name = "hesp-log-archive"
log_archive_bucket_arn  = "arn:aws:s3:::hesp-log-archive"
config_role_arn         = "arn:aws:iam::123456789012:role/OrgConfigRole"
security_contact_email  = "security@hesp.com"
account_alias           = "workloads"
owner                   = "ajamu"
extra_tags = {
  CostCenter = "1234"
}
