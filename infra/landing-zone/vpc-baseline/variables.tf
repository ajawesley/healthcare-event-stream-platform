variable "aws_region" {
  type        = string
  description = "AWS region for the baseline VPC"
}

variable "org_name" {
  type        = string
  description = "Organization or project name (e.g., hesp)"
}

variable "vpc_cidr" {
  type        = string
  description = "CIDR block for the baseline VPC"
  default     = "10.1.0.0/16"
}

variable "azs" {
  type        = list(string)
  description = "List of availability zones to use"
  default     = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

variable "log_archive_bucket_arn" {
  type        = string
  description = "ARN of the centralized log archive bucket"
}

variable "owner" {
  type        = string
  description = "Owner tag value"
}

variable "extra_tags" {
  type        = map(string)
  default     = {}
  description = "Additional tags to apply to all VPC resources"
}
