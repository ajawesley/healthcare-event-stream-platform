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

# Workload Account Provider
provider "aws" {
  region = var.aws_region

  assume_role {
    role_arn = var.workload_account_role_arn
  }
}

############################################
# Locals
############################################

locals {
  base_tags = {
    Project     = "acme"
    Owner       = "security-team"
    Environment = var.environment
  }
}

############################################
# GuardDuty Member (Workload Account)
############################################

module "guardduty_member" {
  source = "../../modules/guardduty_member"

  master_account_id = var.security_admin_account_id
  name_prefix       = var.environment
  tags              = local.base_tags
}

############################################
# Security Hub Member (Workload Account)
############################################

module "securityhub_member" {
  source = "../../modules/securityhub_member"

  admin_account_id = var.security_admin_account_id
  tags             = local.base_tags
}

############################################
# Inspector Member (Workload Account)
############################################

module "inspector_member" {
  source = "../../modules/inspector_member"

  admin_account_id  = var.security_admin_account_id
  member_account_id = var.workload_account_id
  tags              = local.base_tags
}
