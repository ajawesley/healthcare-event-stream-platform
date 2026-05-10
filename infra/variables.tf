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

variable "s3_output_base_path" {
  type = string
}

variable "s3_error_path" {
  type = string
}

# --- NEW: ACM Certificate ARN ---

variable "acm_certificate_arn" {
  type = string
}

# --- RDS PostgreSQL (Compliance DB) ---
variable "compliance_db_username" {
  type        = string
  description = "Username for the compliance PostgreSQL database"
}

variable "compliance_db_host" {
  type        = string
  description = "Hostname or endpoint of the compliance PostgreSQL database"
}

variable "compliance_db_port" {
  type        = number
  description = "Port number for the compliance PostgreSQL database"
}

variable "compliance_db_name" {
  type        = string
  description = "Name of the compliance PostgreSQL database"
}

variable "compliance_db_password" {
  type        = string
  sensitive   = true
  description = "Password for the compliance PostgreSQL database"
}

variable "compliance_db_password_secret_arn" {
  type = string
}

