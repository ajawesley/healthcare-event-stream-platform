############################################################
# Core ECS Service Variables
############################################################

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

variable "kms_key_arn" {
  type = string
}

variable "s3_prefix" {
  type    = string
  default = "events"
}

variable "log_group_name" {
  type = string
}

variable "desired_count" {
  type = number
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

variable "tags" {
  type = map(string)
}

############################################################
# ADOT / OpenTelemetry
############################################################

variable "enable_adot" {
  type    = bool
  default = true
}

# Custom ADOT image (must include /etc/otel/config.yaml)
variable "adot_image" {
  type = string
}

############################################################
# Observability Vendor Keys (Honeycomb)
############################################################

# IMPORTANT:
# This must be a Secrets Manager ARN, not a raw string.
variable "honeycomb_api_key" {
  type      = string
  sensitive = true
}

variable "honeycomb_dataset" {
  type = string
}
