variable "name" {
  description = "Base name for Redis resources (e.g. app-env-compliance-redis)"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID where Redis will be deployed"
  type        = string
}

variable "isolated_subnet_ids" {
  description = "List of isolated subnet IDs for Redis subnet group"
  type        = list(string)
}

variable "ingestion_service_sg_id" {
  description = "Security group ID of the ECS ingestion service"
  type        = string
}

variable "engine_version" {
  description = "Redis engine version"
  type        = string
  default     = "7.0"
}

variable "node_type" {
  description = "Instance type for Redis nodes"
  type        = string
  default     = "cache.t4g.small"
}

variable "port" {
  description = "Redis port"
  type        = number
  default     = 6379
}

variable "parameter_group_family" {
  description = "ElastiCache parameter group family"
  type        = string
  default     = "redis7"
}

variable "maxmemory_policy" {
  description = "Redis maxmemory-policy parameter"
  type        = string
  default     = "allkeys-lru"
}

variable "replicas_per_node_group" {
  description = "Number of replicas per node group (1 = primary + 1 replica)"
  type        = number
  default     = 1
}

variable "transit_encryption_enabled" {
  description = "Enable in-transit encryption for Redis"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags to apply to all Redis resources"
  type        = map(string)
  default     = {}
}
