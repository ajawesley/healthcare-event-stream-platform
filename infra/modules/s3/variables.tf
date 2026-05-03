variable "bucket_name" {
  type        = string
  description = "Name of the HESP raw event S3 bucket. Must match pattern: hesp-{env}-raw-events-{suffix}."

  validation {
    condition     = can(regex("^hesp-[a-z]+-raw-events-[a-z0-9-]+$", var.bucket_name))
    error_message = "bucket_name must match: hesp-{env}-raw-events-{suffix} using lowercase letters, digits, and hyphens only."
  }
}

variable "environment" {
  type        = string
  description = "Deployment environment. Controls Object Lock mode and deletion protection."

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "environment must be one of: dev, staging, prod."
  }
}

variable "owner" {
  type        = string
  description = "Owning team name. Applied as a required tag for cost allocation and incident routing."
}

variable "cost_center" {
  type        = string
  description = "Aetna cost center code. Applied as a required tag for billing allocation."
}

variable "ingest_task_role_arn" {
  type        = string
  description = "ARN of the ECS ingest task IAM role. Granted s3:PutObject on this bucket only."
}

variable "access_log_bucket_id" {
  type        = string
  description = "ID of the centralised audit log S3 bucket. Must exist before this module is called. Server access logs are written here."
}

variable "tags" {
  type        = map(string)
  description = "Additional resource tags. Required platform tags are merged in automatically and cannot be overridden by this map."
  default     = {}
}
