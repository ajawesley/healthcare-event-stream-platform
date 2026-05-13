variable "admin_account_id" {
  type = string
}

variable "member_account_ids" {
  type = list(string)
}

variable "region" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}
