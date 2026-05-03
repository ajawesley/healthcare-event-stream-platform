variable "app_name" {
  type = string
}

variable "environment" {
  type = string
}

variable "cluster_name" {
  type = string
}

variable "container_image" {
  type = string
}

variable "task_execution_role_arn" {
  type = string
}

variable "task_role_arn" {
  type = string
}

variable "subnet_ids" {
  type = list(string)
}

variable "security_group_ids" {
  type = list(string)
}

variable "s3_bucket_name" {
  type = string
}

variable "desired_count" {
  type    = number
  default = 2
}

variable "target_group_arn" {
  type = string
}

variable "owner" {
  type = string
}

variable "cost_center" {
  type = string
}

variable "log_group_name" {
  type        = string
  description = "Name of the CloudWatch log group for ECS task logs."
}

variable "tags" {
  type    = map(string)
  default = {}
}
