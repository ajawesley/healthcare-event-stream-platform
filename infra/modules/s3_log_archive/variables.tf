variable "bucket_name" {
  description = "Name of the log archive S3 bucket"
  type        = string
}

variable "tags" {
  description = "Tags to apply to the log archive bucket"
  type        = map(string)
  default     = {}
}

variable "kms_key_arn" {
  description = "KMS key ARN used for bucket encryption"
  type        = string
}

