############################################
# Create AWS Account
############################################

resource "aws_organizations_account" "this" {
  name      = var.account_name
  email     = var.account_email
  parent_id = var.ou_id

  tags = var.tags
}

############################################
# Provider for the NEW Account
############################################

provider "aws" {
  alias  = "new"
  region = var.aws_region

  assume_role {
    role_arn = "arn:aws:iam::${aws_organizations_account.this.id}:role/${var.bootstrap_role_name}"
  }
}

############################################
# Include Submodules
############################################

module "vpc" {
  source = "./vpc.tf"
  providers = {
    aws = aws.new
  }
}

module "iam" {
  source = "./iam.tf"
  providers = {
    aws = aws.new
  }
}

module "security" {
  source = "./security.tf"
  providers = {
    aws = aws.new
  }

  security_admin_account_id = var.security_admin_account_id
}

module "config" {
  source = "./config.tf"
  providers = {
    aws = aws.new
  }

  bucket_name = var.config_bucket_name
}
