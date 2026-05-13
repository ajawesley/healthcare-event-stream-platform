variable "admin_account_id" {
  description = "Inspector delegated admin account ID"
  type        = string
}

variable "member_account_id" {
  description = "This workload account ID"
  type        = string
}

variable "tags" {
  description = "Tags to apply to Inspector member resources"
  type        = map(string)
  default     = {}
}
