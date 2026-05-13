variable "name_prefix" {
  description = "Prefix for naming AWS Config resources"
  type        = string
}

variable "region" {
  description = "AWS region"
  type        = string
}

variable "log_archive_bucket_arn" {
  description = "S3 bucket ARN where AWS Config delivers snapshots"
  type        = string
}

variable "log_archive_bucket_name" {
  description = "S3 bucket name for AWS Config delivery channel"
  type        = string
}

variable "kms_key_arn" {
  description = "KMS key ARN used for encrypting AWS Config data"
  type        = string
}

variable "config_role_arn" {
  description = "IAM role ARN for AWS Config Recorder"
  type        = string
}

variable "tags" {
  description = "Tags to apply to AWS Config resources"
  type        = map(string)
  default     = {}
}
