variable "account_id" {
  description = "Delegated admin account ID for Inspector"
  type        = string
}

variable "tags" {
  description = "Tags to apply to Inspector delegated admin resources"
  type        = map(string)
  default     = {}
}
