variable "app_name" {
  type = string
}

variable "environment" {
  type = string
}

variable "owner" {
  type = string
}

variable "cost_center" {
  type = string
}

variable "tags" {
  type = map(string)
}

variable "glue_role_arn" {
  type = string
}

variable "script_s3_path" {
  type = string
}

variable "temp_dir" {
  type = string
}

variable "log_group_name" {
  type = string
}

variable "kms_key_arn" {
  type      = string
  default   = null
  nullable  = true
}
