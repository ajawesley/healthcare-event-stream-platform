############################################
# Terraform + Providers
############################################

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Delegated Admin (Security Account) Provider
provider "aws" {
  region = var.aws_region

  assume_role {
    role_arn = var.security_admin_role_arn
  }
}

############################################
# Locals
############################################

locals {
  base_tags = {
    Project = "acme"
    Owner   = "security-team"
  }
}

############################################
# Inspector Delegated Admin (Security Account)
############################################

module "inspector_admin" {
  source = "../../modules/inspector_admin"

  account_id = var.security_admin_account_id
  tags       = local.base_tags
}

############################################
# GuardDuty Delegated Admin (Security Account)
############################################

module "guardduty_admin" {
  source = "../../modules/guardduty_admin"

  name_prefix = var.org_name
  tags        = local.base_tags
}

############################################
# Security Hub Delegated Admin (Security Account)
############################################

module "securityhub_admin" {
  source = "../../modules/securityhub_admin"

  account_id = var.security_admin_account_id
  tags       = local.base_tags
}
