############################################
# S3 Bucket Module Variables
############################################

variable "raw_bucket_name" {
  description = "Name of the raw events S3 bucket"
  type        = string
}

variable "golden_bucket_name" {
  description = "Name of the golden events S3 bucket"
  type        = string
}

variable "script_bucket_name" {
  description = "Name of the Glue scripts S3 bucket"
  type        = string
}

variable "access_logs_bucket_name" {
  description = "Name of the access logs S3 bucket"
  type        = string
}

variable "log_archive_bucket_name" {
  description = "Name of the CloudTrail/Config/GuardDuty log archive bucket"
  type        = string
}

variable "kms_key_arn" {
  description = "KMS key ARN used for bucket encryption"
  type        = string
}

variable "error_prefix" {
  description = "Prefix for golden bucket error folder"
  type        = string
}

variable "tags" {
  description = "Common tags applied to all S3 buckets"
  type        = map(string)
}
