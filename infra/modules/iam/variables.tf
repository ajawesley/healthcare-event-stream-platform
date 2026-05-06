variable "environment" {
  type = string
}

variable "owner" {
  type = string
}

variable "cost_center" {
  type = string
}

variable "kms_key_arn" {
  type = string
}

variable "log_group_arn" {
  type = string
}

variable "tags" {
  type = map(string)
}

variable "raw_bucket_arn" {
  type = string
}

variable "script_bucket_arn" {
  type = string
}

variable "golden_bucket_arn" {
  type = string
}

