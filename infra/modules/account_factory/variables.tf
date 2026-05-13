variable "account_name" {
  type = string
}

variable "account_email" {
  type = string
}

variable "ou_id" {
  type = string
}

variable "aws_region" {
  type = string
}

variable "bootstrap_role_name" {
  type    = string
  default = "OrganizationAccountAccessRole"
}

variable "security_admin_account_id" {
  type = string
}

variable "config_bucket_name" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}
