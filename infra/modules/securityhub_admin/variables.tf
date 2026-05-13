variable "account_id" {
  description = "Delegated admin account ID for Security Hub"
  type        = string
}

variable "tags" {
  description = "Tags to apply to Security Hub delegated admin resources"
  type        = map(string)
  default     = {}
}
