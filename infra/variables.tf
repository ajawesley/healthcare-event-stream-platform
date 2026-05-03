variable "aws_region" {
  type        = string
  description = "AWS region to deploy resources into."
}

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
  description = "Owner tag value."
}

variable "cost_center" {
  type        = string
  description = "Cost center tag value."
}

variable "bucket_name" {
  type        = string
  description = "Name of the S3 bucket for ingest raw events."
}

variable "access_log_bucket_name" {
  type        = string
  description = "Name of the central access logs bucket."
}

variable "vpc_id" {
  type        = string
  description = "VPC ID for ALB and ECS."
}

variable "public_subnet_ids" {
  type        = list(string)
  description = "Public subnets for ALB."
}

variable "private_subnet_ids" {
  type        = list(string)
  description = "Private subnets for ECS tasks."
}

variable "container_image" {
  type        = string
  description = "ECS container image."
}

variable "desired_count" {
  type        = number
  description = "Desired ECS task count."
  default     = 2
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "Additional tags."
}
