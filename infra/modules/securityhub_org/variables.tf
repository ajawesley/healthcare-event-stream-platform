variable "admin_account_id" {
  description = "Delegated admin account for Security Hub"
  type        = string
}

variable "tags" {
  description = "Tags to apply to Security Hub org resources"
  type        = map(string)
  default     = {}
}
