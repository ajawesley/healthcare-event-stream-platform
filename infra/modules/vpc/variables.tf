variable "name_prefix" {
  description = "Prefix used for naming VPC and subnet resources (e.g., acme-dev)"
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.10.0.0/16"
}

variable "az_count" {
  description = "Number of AZs to use (recommended: 3)"
  type        = number
  default     = 3
}

variable "tags" {
  description = "Common tags to apply to all VPC resources"
  type        = map(string)
  default     = {}
}
