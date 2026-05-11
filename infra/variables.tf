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

# --- ACM Certificate ARN ---
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

# --- DynamoDB Compliance Rules Table ---

variable "dynamodb_ttl_enabled" {
  type    = bool
  default = false
}

variable "dynamodb_ttl_attribute_name" {
  type    = string
  default = "expires_at"
}

variable "dynamodb_pitr_enabled" {
  type    = bool
  default = true
}

# --- Redis Compliance Cache (Multi-AZ) ---
variable "redis_name" {
  type = string
}

variable "redis_engine_version" {
  type    = string
  default = "7.0"
}

variable "redis_node_type" {
  type    = string
  default = "cache.t4g.small"
}

variable "redis_port" {
  type    = number
  default = 6379
}

variable "redis_parameter_group_family" {
  type    = string
  default = "redis7"
}

variable "redis_maxmemory_policy" {
  type    = string
  default = "allkeys-lru"
}

variable "redis_replicas_per_node_group" {
  type    = number
  default = 1
}

variable "redis_transit_encryption_enabled" {
  type    = bool
  default = true
}

