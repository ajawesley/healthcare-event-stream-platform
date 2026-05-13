locals {
  aws_region = "us-east-1"
  org_name   = "hesp"
  owner      = "ajamu"
}

remote_state {
  backend = "s3"
  config = {
    bucket         = "hesp-landing-zone-tfstate"
    key            = "${path_relative_to_include()}/terraform.tfstate"
    region         = local.aws_region
    encrypt        = true
    dynamodb_table = "hesp-landing-zone-locks"
  }
}
