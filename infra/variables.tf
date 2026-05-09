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

variable "container_image" {
  type = string
}

variable "desired_count" {
  type = number
}

variable "script_bucket" {
  type = string
}

variable "glue_script_s3_path" {
  type = string
}

variable "glue_temp_dir" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}

# --- Observability Vendor Keys (Secrets Manager ARNs) ---

variable "honeycomb_api_key" {
  type      = string
  sensitive = true
}

variable "honeycomb_dataset" {
  type = string
}

# --- NEW: ACM Certificate ARN ---

variable "acm_certificate_arn" {
  type = string
}
