############################################
# RDS PostgreSQL Variables
############################################

variable "name" {
  description = "Name prefix / identifier for the RDS instance (e.g. app-env-compliance-db)"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID where the Compliance DB will be deployed"
  type        = string
}

variable "isolated_subnet_ids" {
  description = "List of isolated subnet IDs for the DB subnet group"
  type        = list(string)
}

variable "ingestion_service_sg_id" {
  description = "Security group ID for the ECS ingestion service"
  type        = string
}

variable "instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

variable "engine_version" {
  description = "PostgreSQL engine version"
  type        = string
  default     = "14.22"
}

variable "allocated_storage" {
  description = "Initial allocated storage (GB)"
  type        = number
  default     = 20
}

variable "max_allocated_storage" {
  description = "Maximum autoscaled storage (GB)"
  type        = number
  default     = 100
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "compliance"
}

variable "db_username" {
  description = "Master username for the Compliance DB"
  type        = string
}

variable "db_password" {
  description = "Master password for the Compliance DB"
  type        = string
  sensitive   = true

  validation {
    condition = (
      length(var.db_password) >= 8 &&
      can(regex("^[A-Za-z0-9!#$%^&*()_+=\\-{}\\[\\]:;,.?]+$", var.db_password))
    )
    error_message = "RDS password contains invalid characters. Allowed: letters, numbers, and !#$%^&*()_+=-{}[]:;,.?"
  }
}

variable "multi_az" {
  description = "Enable Multi-AZ deployment"
  type        = bool
  default     = false
}

variable "backup_retention_days" {
  description = "Backup retention period in days"
  type        = number
  default     = 7
}

variable "tags" {
  description = "Common tags to apply to all Compliance DB resources"
  type        = map(string)
  default     = {}
}
