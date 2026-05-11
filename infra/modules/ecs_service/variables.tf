############################################################
# Core ECS Service Variables
############################################################

variable "app_name" { type = string }
variable "environment" { type = string }
variable "cluster_name" { type = string }
variable "container_image" { type = string }

variable "task_execution_role_arn" { type = string }
variable "task_role_arn" { type = string }

variable "subnet_ids" { type = list(string) }
variable "security_group_ids" { type = list(string) }

variable "s3_bucket_name" { type = string }
variable "kms_key_arn" { type = string }
variable "s3_prefix" {
  type    = string
  default = "events"
}

variable "log_group_name" { type = string }
variable "desired_count" { type = number }
variable "target_group_arn" { type = string }

variable "owner" { type = string }
variable "cost_center" { type = string }
variable "tags" { type = map(string) }

############################################################
# ADOT / OpenTelemetry
############################################################

variable "enable_adot" {
  type    = bool
  default = true
}

variable "adot_image" { type = string }

############################################################
# Compliance DB Wiring
############################################################

variable "compliance_db_host" { type = string }
variable "compliance_db_port" { type = number }
variable "compliance_db_name" { type = string }
variable "compliance_db_username" { type = string }

variable "compliance_db_password_secret_arn" {
  type      = string
  sensitive = true
}

############################################################
# DynamoDB Compliance Table
############################################################

variable "dynamodb_table_name" {
  type = string
}

############################################################
# Redis Compliance Cache
############################################################

variable "redis_primary_endpoint" {
  type = string
}
