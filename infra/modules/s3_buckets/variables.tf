############################################
# Core Metadata
############################################

variable "app_name" {
  description = "Application name used for naming S3 buckets"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, qa, prod)"
  type        = string
}

variable "aws_region" {
  description = "AWS region where buckets will be created"
  type        = string
}

variable "owner" {
  description = "Owner tag for cost allocation"
  type        = string
}

variable "cost_center" {
  description = "Cost center tag for billing"
  type        = string
}

variable "tags" {
  description = "Base tags applied to all S3 buckets"
  type        = map(string)
}

############################################
# Bucket Names
############################################

variable "raw_bucket_name" {
  description = "Name of the raw events S3 bucket"
  type        = string
}

variable "access_logs_bucket_name" {
  description = "Name of the access logs bucket"
  type        = string
}

variable "golden_bucket_name" {
  description = "Name of the golden events bucket"
  type        = string
}

variable "script_bucket_name" {
  description = "Name of the Glue scripts bucket"
  type        = string
}

############################################
# Glue + ETL Paths
############################################

variable "glue_script_s3_path" {
  description = "S3 path for Glue ETL scripts"
  type        = string
}

variable "glue_temp_dir" {
  description = "Temporary directory for Glue jobs"
  type        = string
}

############################################
# Error Prefix
############################################

variable "error_prefix" {
  description = "Prefix for error objects in the golden bucket"
  type        = string
}

############################################
# Output + Error Paths
############################################

variable "s3_output_base_path" {
  description = "Base S3 path for ETL output"
  type        = string
}

variable "s3_error_path" {
  description = "S3 path for ETL error output"
  type        = string
}

############################################
# Lambda Packaging
############################################

variable "lambda_zip_path" {
  description = "Path to the Lambda ZIP file for ingestion"
  type        = string
}

############################################
# Encryption
############################################

variable "kms_key_arn" {
  description = "KMS key ARN used for SSE-KMS encryption"
  type        = string
}
