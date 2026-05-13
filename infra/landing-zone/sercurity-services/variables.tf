variable "aws_region" {
  type        = string
  description = "AWS region for the security services account"
}

variable "org_name" {
  type        = string
  description = "Organization or project name (e.g., hesp)"
}

variable "security_account_id" {
  type        = string
  description = "AWS Account ID of the Security account (delegated admin)"
}

variable "owner" {
  type        = string
  description = "Owner tag value"
}

variable "extra_tags" {
  type        = map(string)
  default     = {}
  description = "Additional tags to apply to all security services resources"
}
