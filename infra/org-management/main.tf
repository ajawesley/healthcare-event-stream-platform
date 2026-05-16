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

provider "aws" {
  region = var.aws_region

  assume_role {
    role_arn = var.org_management_role_arn
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
# OU Structure (Root → Security, Infra, Workloads, Sandbox)
############################################

module "ou_structure" {
  source = "../modules/ou_structure"

  root_id = var.org_root_id
}

############################################
# Log Archive Bucket (ORG-level)
############################################

module "s3_log_archive" {
  source = "../modules/s3_log_archive"

  bucket_name = "${var.org_name}-log-archive"
  tags        = local.base_tags

  # NEW: required for KMS-encrypted log archive bucket
  kms_key_arn = module.kms_cloudtrail.arn
}

############################################
# CloudTrail KMS Key (ORG-level)
############################################

module "kms_cloudtrail" {
  source = "../modules/kms_cloudtrail"

  name_prefix = var.org_name
  tags        = local.base_tags
}

############################################
# Organization-wide CloudTrail
############################################

module "cloudtrail_org" {
  source = "../modules/cloudtrail_org"

  name_prefix             = var.org_name
  log_archive_bucket_name = module.s3_log_archive.bucket_name
  kms_key_arn             = module.kms_cloudtrail.arn

  tags = local.base_tags
}

############################################
# Security Hub ORG-level (Management Account)
############################################

module "securityhub_org" {
  source = "../modules/securityhub_org"

  admin_account_id = var.security_admin_account_id
  tags             = local.base_tags
}

############################################
# Inspector ORG-level (Management Account)
############################################

module "inspector_org" {
  source = "../modules/inspector_org"

  admin_account_id = var.security_admin_account_id
  tags             = local.base_tags
}

############################################
# IAM Baseline (ORG-level)
############################################

module "iam_org_baseline" {
  source = "../modules/iam_org_baseline"

  security_admin_account_id = var.security_admin_account_id
  tags                      = local.base_tags
}

############################################
# SCP Baseline (Attach to Organization Root)
############################################

module "scp_baseline" {
  source = "../modules/scp_baseline"

  scp_directory = "${path.module}/../../landing-zone/org/scp"
  root_id       = var.org_root_id
}

############################################
# IAM Role for AWS Config Aggregator
############################################

module "config_aggregation_role" {
  source = "../modules/config_aggregation_role"

  name_prefix = var.org_name
  tags        = local.base_tags
}

############################################
# AWS Config Organization Aggregator
############################################

module "config_aggregator" {
  source = "../modules/config_aggregator"

  name                 = "${var.org_name}-org-config-aggregator"
  aggregation_role_arn = module.config_aggregation_role.role_arn
  all_regions          = true
  tags                 = local.base_tags
}

############################################
# Centralized Logging (CloudWatch → S3 → Athena)
############################################

module "centralized_logging" {
  source = "../modules/centralized_logging"

  name_prefix             = var.org_name
  log_archive_bucket_name = module.s3_log_archive.bucket_name
  log_archive_bucket_arn  = module.s3_log_archive.bucket_arn

  log_group_names = [
    "/aws/lambda/ingestion-service",
    "/aws/lambda/compliance-service",
    "/aws/ecs/ingestion-service"
  ]

  lambda_zip_path = "${path.module}/lambda/cw_forwarder.zip"

  aws_region = var.aws_region
  account_id = var.security_admin_account_id

  tags = local.base_tags
}

############################################
# Account Placement (Move existing accounts into OUs)
############################################

module "place_dev_account" {
  source = "../modules/account_placement"

  account_id    = "111111111111"
  account_name  = "acme-dev"
  account_email = "acme-dev+aws@example.com"

  ou_id = module.ou_structure.ou_ids["dev"]

  tags = local.base_tags
}

module "place_qa_account" {
  source = "../modules/account_placement"

  account_id    = "222222222222"
  account_name  = "acme-qa"
  account_email = "acme-qa+aws@example.com"

  ou_id = module.ou_structure.ou_ids["qa"]

  tags = local.base_tags
}

module "place_prod_account" {
  source = "../modules/account_placement"

  account_id    = "333333333333"
  account_name  = "acme-prod"
  account_email = "acme-prod+aws@example.com"

  ou_id = module.ou_structure.ou_ids["prod"]

  tags = local.base_tags
}
