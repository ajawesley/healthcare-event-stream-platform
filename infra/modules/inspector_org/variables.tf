variable "admin_account_id" {
  description = "Delegated admin account for Inspector"
  type        = string
}

variable "tags" {
  description = "Tags to apply to Inspector org resources"
  type        = map(string)
  default     = {}
}
