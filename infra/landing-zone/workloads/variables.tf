variable "aws_region" {
  type        = string
  description = "AWS region for the workloads account"
}

variable "org_name" {
  type        = string
  description = "Organization or project name (e.g., hesp)"
}

variable "workloads_account_id" {
  type        = string
  description = "AWS Account ID of the workloads account"
}

variable "log_archive_bucket_name" {
  type        = string
  description = "Name of the org-level log archive bucket"
}

variable "log_archive_bucket_arn" {
  type        = string
  description = "ARN of the org-level log archive bucket"
}

variable "config_role_arn" {
  type        = string
  description = "IAM role ARN used by AWS Config recorder"
}

variable "security_contact_email" {
  type        = string
  description = "Email address for GuardDuty/SecurityHub invitations"
}

variable "account_alias" {
  type        = string
  description = "Friendly alias for the workloads account"
}

variable "vpc_cidr" {
  type    = string
  default = "10.1.0.0/16"
}

variable "azs" {
  type    = list(string)
  default = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

variable "owner" {
  type = string
}

variable "extra_tags" {
  type    = map(string)
  default = {}
}
