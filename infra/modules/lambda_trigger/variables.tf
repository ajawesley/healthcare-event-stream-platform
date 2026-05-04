############################################
# Lambda Trigger Module Variables
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
  description = "Additional tags to apply to the Lambda function."
}

variable "glue_job_name" {
  type        = string
  description = "Name of the Glue job to trigger."
}

variable "glue_job_arn" {
  type        = string
  description = "ARN of the Glue job to trigger."
}

variable "lambda_zip_path" {
  type        = string
  description = "Path to the built Lambda ZIP file."
}
