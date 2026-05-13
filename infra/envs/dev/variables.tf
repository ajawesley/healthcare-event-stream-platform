############################################
# Core App + Environment
############################################

variable "app_name" {
  description = "Application name"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, qa, prod)"
  type        = string
}

variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
}

############################################
# Org-Level GitHub OIDC Deploy Role
############################################

variable "github_oidc_role_arn" {
  description = "ARN of the org-level GitHub OIDC deploy role"
  type        = string
}

############################################
# Ownership + Cost
############################################

variable "owner" {
  description = "Owner of the application"
  type        = string
}

variable "cost_center" {
  description = "Cost center for billing"
  type        = string
}

############################################
# S3 Buckets
############################################

variable "raw_bucket_name" {
  description = "Name of the raw events S3 bucket"
  type        = string
}

variable "access_logs_bucket_name" {
  description = "Name of the access logs S3 bucket"
  type        = string
}

variable "s3_output_base_path" {
  description = "Base path for S3 outputs"
  type        = string
}

############################################
# ECS + Container
############################################

variable "container_image" {
  description = "Container image for ECS service"
  type        = string
}

variable "desired_count" {
  description = "Desired ECS task count"
  type        = number
  default     = 1
}

variable "adot_image" {
  description = "ADOT container image"
  type        = string
}

############################################
# Compliance DB
############################################

variable "compliance_db_password" {
  description = "Password for the compliance DB"
  type        = string
  sensitive   = true
}

variable "compliance_db_name" {
  description = "Name of the compliance DB"
  type        = string
}

variable "compliance_db_username" {
  description = "Username for the compliance DB"
  type        = string
}

variable "compliance_db_password_secret_arn" {
  description = "ARN of the compliance DB password secret"
  type        = string
  sensitive   = true
}

############################################
# Glue
############################################

variable "glue_script_s3_path" {
  description = "S3 path for the Glue job script"
  type        = string
}

variable "glue_temp_dir" {
  description = "Temporary directory for Glue jobs"
  type        = string
}

############################################
# Lambda
############################################

variable "lambda_zip_path" {
  description = "Path to the packaged Lambda ZIP file"
  type        = string

}

variable "workload_account_id" {
  type = string
}

variable "s3_error_path" {
  type = string
}

variable "log_archive_bucket_name" {
  type = string
}

