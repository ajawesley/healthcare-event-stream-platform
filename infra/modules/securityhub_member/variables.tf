variable "admin_account_id" {
  description = "Security Hub delegated admin account ID"
  type        = string
}

variable "tags" {
  description = "Tags to apply to Security Hub member resources"
  type        = map(string)
  default     = {}
}
