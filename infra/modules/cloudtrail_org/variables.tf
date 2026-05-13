variable "name_prefix" {
  description = "Prefix for CloudTrail naming"
  type        = string
}

variable "log_archive_bucket_name" {
  description = "S3 bucket name for CloudTrail logs"
  type        = string
}

variable "kms_key_arn" {
  description = "KMS key ARN for encrypting CloudTrail logs"
  type        = string
}

variable "is_organization_trail" {
  description = "Whether this is an organization-wide trail"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags to apply to CloudTrail resources"
  type        = map(string)
  default     = {}
}
