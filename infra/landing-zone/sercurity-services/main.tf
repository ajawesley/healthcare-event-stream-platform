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
      Owner   = var.owner
    },
    var.extra_tags
  )
}

############################################
# GuardDuty Delegated Administrator
############################################

resource "aws_guardduty_organization_admin_account" "this" {
  admin_account_id = var.security_account_id
}

resource "aws_guardduty_detector" "security_admin" {
  provider = aws
  enable   = true

  tags = local.tags
}

############################################
# Security Hub Delegated Administrator
############################################

resource "aws_securityhub_organization_admin_account" "this" {
  admin_account_id = var.security_account_id
}

resource "aws_securityhub_account" "security_admin" {
  provider = aws

  tags = local.tags
}

############################################
# IAM Access Analyzer (Organization)
############################################

resource "aws_accessanalyzer_analyzer" "org" {
  analyzer_name = "${var.org_name}-org-analyzer"
  type          = "ORGANIZATION"

  tags = local.tags
}

############################################
# Organization-wide Security Hub Standards
############################################

resource "aws_securityhub_standards_subscription" "cis" {
  standards_arn = "arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
}

resource "aws_securityhub_standards_subscription" "aws_foundational" {
  standards_arn = "arn:aws:securityhub:::ruleset/aws-foundational-security-best-practices/v/1.0.0"
}

############################################
# Organization-wide GuardDuty Settings
############################################

resource "aws_guardduty_organization_configuration" "this" {
  detector_id = aws_guardduty_detector.security_admin.id

  auto_enable_organization_members = "ALL"
}
