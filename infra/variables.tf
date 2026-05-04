variable "aws_region" {
  type = string
}

variable "app_name" {
  type = string
}

variable "environment" {
  type = string
}

variable "owner" {
  type = string
}

variable "cost_center" {
  type = string
}

variable "bucket_name" {
  type = string
}

variable "access_log_bucket_name" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "public_subnet_ids" {
  type = list(string)
}

variable "private_subnet_ids" {
  type = list(string)
}

variable "container_image" {
  type = string
}

variable "desired_count" {
  type = number
}

variable "account_id" {
  type = string
}

variable "enable_schedule" {
  type    = bool
  default = false
}

variable "schedule_expression" {
  type    = string
  default = "rate(1 day)"
}
