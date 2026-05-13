############################################
# ORG Management Variables
############################################

variable "aws_region" {
  description = "AWS region for ORG management operations"
  type        = string
}

variable "org_management_role_arn" {
  description = "Role ARN to assume into the ORG management account"
  type        = string
}

variable "org_name" {
  description = "Organization or project name prefix"
  type        = string
}

variable "security_admin_account_id" {
  description = "Delegated admin (security) account ID"
  type        = string
}

variable "org_root_id" {
  description = "AWS Organizations Root ID (e.g., r-1234)"
  type        = string
}

############################################
# AWS Config Aggregator
############################################

variable "config_aggregation_role_arn" {
  description = "IAM role ARN used by AWS Config aggregator"
  type        = string
}
