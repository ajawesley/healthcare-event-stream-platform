variable "environment" {
  type = string
}

variable "owner" {
  type = string
}

variable "cost_center" {
  type = string
}

variable "bucket_arn" {
  type = string
}

variable "kms_key_arn" {
  type = string
}

variable "log_group_arn" {
  type = string
}

variable "scripts_bucket_arn" {
  type        = string
  description = "ARN of the S3 bucket containing Glue scripts and Python libraries."
}

variable "tags" {
  type    = map(string)
  default = {}
}
