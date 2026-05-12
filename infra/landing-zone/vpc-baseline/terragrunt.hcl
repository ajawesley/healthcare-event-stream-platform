include {
  path = find_in_parent_folders()
}

terraform {
  source = "../vpc-baseline"
}

inputs = {
  aws_region             = "us-east-1"
  org_name               = "hesp"
  vpc_cidr               = "10.1.0.0/16"
  log_archive_bucket_arn = "arn:aws:s3:::hesp-log-archive"
  owner                  = "ajamu"
}
