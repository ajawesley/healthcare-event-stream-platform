variable "aws_region" {
  type        = string
  description = "AWS region for the logging account"
}

variable "org_name" {
  type        = string
  description = "Organization or project name (e.g., hesp)"
}

variable "log_archive_bucket_name" {
  type        = string
  description = "Name of the centralized log archive bucket"
}

variable "org_config_role_arn" {
  type        = string
  description = "IAM role ARN used by AWS Config aggregator at the org level"
}

variable "owner" {
  type        = string
  description = "Owner tag value"
}

variable "extra_tags" {
  type        = map(string)
  default     = {}
  description = "Additional tags to apply to all logging resources"
}
