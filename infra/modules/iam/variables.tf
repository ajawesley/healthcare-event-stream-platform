variable "environment" {
  type        = string
  description = "Deployment environment."
}

variable "owner" {
  type        = string
  description = "Owner tag."
}

variable "cost_center" {
  type        = string
  description = "Cost center tag."
}

variable "bucket_arn" {
  type        = string
  description = "ARN of the S3 bucket used by ECS and Glue."
}

variable "kms_key_arn" {
  type        = string
  description = "KMS key used for S3 encryption."
}

variable "log_group_arn" {
  type        = string
  description = "CloudWatch log group ARN for ECS tasks."
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "Additional tags."
}

