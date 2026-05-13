variable "account_id" {
  description = "AWS Account ID to move into the OU"
  type        = string
}

variable "account_name" {
  description = "Account name (must match existing account)"
  type        = string
}

variable "account_email" {
  description = "Account email (must match existing account)"
  type        = string
}

variable "ou_id" {
  description = "Target OU ID"
  type        = string
}

variable "tags" {
  description = "Tags to apply to the account"
  type        = map(string)
  default     = {}
}
