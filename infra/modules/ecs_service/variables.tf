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

# --- ADOT / OTEL ---

variable "enable_adot" {
  type    = bool
  default = true
}

variable "adot_image" {
  type    = string
  default = "public.ecr.aws/aws-observability/aws-otel-collector:latest"
}

variable "adot_config_file" {
  type = string
}

# --- Observability Vendor Keys ---

variable "dd_api_key" {
  type      = string
  default   = ""
  sensitive = true
}

variable "honeycomb_api_key" {
  type      = string
  default   = ""
  sensitive = true
}

variable "honeycomb_dataset" {
  type    = string
  default = ""
}
