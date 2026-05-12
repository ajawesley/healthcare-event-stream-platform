include {
  path = find_in_parent_folders()
}

terraform {
  source = "../org"
}

inputs = {
  aws_region = "us-east-1"
  org_name   = "hesp"
  owner      = "ajamu"
}
