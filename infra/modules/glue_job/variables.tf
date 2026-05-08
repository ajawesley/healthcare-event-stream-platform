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

variable "script_bucket" {
  description = "S3 bucket containing Glue ETL scripts"
  type        = string
}

variable "script_s3_path" {
  description = "Full S3 path to the Glue job script"
  type        = string
}

variable "temp_dir" {
  description = "S3 temp directory for Glue job"
  type        = string
}

variable "log_group_name" {
  type = string
}

variable "kms_key_arn" {
  type     = string
  default  = null
  nullable = true
}

# Required for Glue arguments
variable "raw_bucket" {
  type = string
}

variable "golden_bucket" {
  type = string
}
