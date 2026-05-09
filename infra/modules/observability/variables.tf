variable "environment" {
  description = "Deployment environment (dev, qa, prod)"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "app_name" {
  description = "Application name"
  type        = string
}

variable "tags" {
  description = "Base tags"
  type        = map(string)
  default     = {}
}

variable "ecs_cluster_name" {
  description = "ECS cluster name for Container Insights metrics"
  type        = string
}

variable "ecs_service_name" {
  description = "ECS service name for Container Insights metrics"
  type        = string
}

variable "alb_arn" {
  description = "Full ARN of the ALB"
  type        = string
}

variable "target_group_arn" {
  description = "Full ARN of the ALB target group"
  type        = string
}

variable "raw_bucket_arn" {
  description = "ARN of the raw events S3 bucket"
  type        = string
}

variable "access_logs_bucket_name" {
  description = "Name of the access logs bucket for CloudTrail"
  type        = string
}
