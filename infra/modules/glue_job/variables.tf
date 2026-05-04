############################################
# Glue Job Module Variables
############################################

variable "app_name" {
  type        = string
  description = "Application name used for naming resources."
}

variable "environment" {
  type        = string
  description = "Deployment environment (dev, staging, prod)."
}

variable "owner" {
  type        = string
  description = "Owning team or individual for tagging."
}

variable "cost_center" {
  type        = string
  description = "Cost center code for tagging and billing."
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "Additional tags to apply to the Glue job."
}

variable "glue_role_arn" {
  type        = string
  description = "IAM role ARN used by the Glue job."
}

variable "script_s3_path" {
  type        = string
  description = "S3 path to the Glue ETL script."
}

variable "temp_dir" {
  type        = string
  description = "S3 path for Glue temporary directory."
}
