variable "security_admin_account_id" {
  description = "Delegated admin (security) account ID"
  type        = string
}

variable "tags" {
  description = "Tags to apply to IAM baseline resources"
  type        = map(string)
  default     = {}
}
