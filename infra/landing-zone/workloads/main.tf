terraform {
  required_version = ">= 1.5.0"
}

provider "aws" {
  region = var.aws_region
}

locals {
  tags = merge(
    {
      Project = var.org_name
      Environment = "workloads"
      Owner = var.owner
    },
    var.extra_tags
  )
}

############################################
# Account-Level CloudTrail
############################################

resource "aws_cloudtrail" "account_trail" {
  name                          = "${var.account_alias}-trail"
  s3_bucket_name                = var.log_archive_bucket_name
  include_global_service_events = true
  is_multi_region_trail         = true
  enable_log_file_validation    = true

  tags = local.tags
}

############################################
# AWS Config (Account-Level)
############################################

resource "aws_config_configuration_recorder" "this" {
  name     = "default"
  role_arn = var.config_role_arn

  recording_group {
    all_supported                 = true
    include_global_resource_types = true
  }
}

resource "aws_config_delivery_channel" "this" {
  name           = "default"
  s3_bucket_name = var.log_archive_bucket_name
}

############################################
# GuardDuty Member
############################################

resource "aws_guardduty_detector" "this" {
  enable = true
}

resource "aws_guardduty_member" "this" {
  account_id               = var.workloads_account_id
  detector_id              = aws_guardduty_detector.this.id
  email                    = var.security_contact_email
  invite                   = true
  disable_email_notification = true
}

############################################
# Security Hub Member
############################################

resource "aws_securityhub_account" "this" {}

resource "aws_securityhub_member" "this" {
  account_id = var.workloads_account_id
  email      = var.security_contact_email
  invite     = true
}

############################################
# IAM Access Analyzer
############################################

resource "aws_accessanalyzer_analyzer" "account" {
  analyzer_name = "${var.account_alias}-analyzer"
  type          = "ACCOUNT"

  tags = local.tags
}

############################################
# Baseline VPC (10.1.0.0/16)
############################################

module "vpc_baseline" {
  source = "../vpc-baseline"

  aws_region             = var.aws_region
  org_name               = var.org_name
  vpc_cidr               = var.vpc_cidr
  azs                    = var.azs
  log_archive_bucket_arn = var.log_archive_bucket_arn
  owner                  = var.owner
  extra_tags             = var.extra_tags
}