variable "aws_region" {
  type = string
}

variable "workload_account_role_arn" {
  description = "Role ARN to assume into this workload account"
  type        = string
}

variable "security_admin_account_id" {
  description = "Delegated admin (security) account ID"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, prod, etc.)"
  type        = string
}

variable "workload_account_id" {
  description = "Workload account ID"
  type        = string
}
