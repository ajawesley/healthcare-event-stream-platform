variable "app_name" {
  type        = string
  description = "Application name used as a prefix for all ALB resource names. Must be lowercase with hyphens only."

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]+$", var.app_name))
    error_message = "app_name must start with a lowercase letter and contain only lowercase letters, digits, and hyphens."
  }
}

variable "environment" {
  type        = string
  description = "Deployment environment. Controls deletion protection and required tags."

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "environment must be one of: dev, staging, prod."
  }
}

variable "vpc_id" {
  type        = string
  description = "ID of the VPC in which the ALB and its security group are created."
}

variable "subnet_ids" {
  type        = list(string)
  description = "List of public subnet IDs across at least two availability zones."

  validation {
    condition     = length(var.subnet_ids) >= 2
    error_message = "At least two subnet IDs are required for multi-AZ ALB deployment."
  }
}

variable "ecs_security_group_id" {
  type        = string
  description = "Security group ID of the ECS ingest service tasks."
}

variable "access_log_bucket_id" {
  type        = string
  description = "ID of the centralised audit log S3 bucket."
}

variable "owner" {
  type        = string
  description = "Owning team name."
}

variable "cost_center" {
  type        = string
  description = "Aetna cost center code."
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "Additional resource tags."
}
