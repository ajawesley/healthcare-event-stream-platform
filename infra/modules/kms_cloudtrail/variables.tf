variable "name_prefix" {
  description = "Prefix for the CloudTrail KMS key alias"
  type        = string
}

variable "tags" {
  description = "Tags to apply to the KMS key"
  type        = map(string)
  default     = {}
}
