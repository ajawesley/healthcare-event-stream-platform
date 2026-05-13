variable "aws_region" {
  type = string
}

variable "security_admin_role_arn" {
  description = "Role ARN to assume into the delegated admin (security) account"
  type        = string
}

variable "security_admin_account_id" {
  description = "Account ID of the delegated admin (security) account"
  type        = string
}

variable "org_name" {
  description = "Organization or project name prefix"
  type        = string
}
