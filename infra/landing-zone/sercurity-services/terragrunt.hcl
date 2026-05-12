include {
  path = find_in_parent_folders()
}

terraform {
  source = "../security-services"
}

inputs = {
  aws_region          = "us-east-1"
  org_name            = "hesp"
  security_account_id = "SECURITY_ACCOUNT_ID"
  owner               = "ajamu"
}
