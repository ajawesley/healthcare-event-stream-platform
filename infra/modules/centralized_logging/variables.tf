variable "name_prefix" {
  type        = string
  description = "Prefix for naming centralized logging resources"
}

variable "log_archive_bucket_name" {
  type        = string
  description = "Name of the centralized log archive bucket"
}

variable "log_archive_bucket_arn" {
  type        = string
  description = "ARN of the centralized log archive bucket"
}

variable "log_group_names" {
  type        = list(string)
  description = "List of CloudWatch log groups to forward to S3"
}

variable "lambda_zip_path" {
  type        = string
  description = "Path to the packaged Lambda ZIP file"
}

variable "aws_region" {
  type = string
}

variable "account_id" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}
