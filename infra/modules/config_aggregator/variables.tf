variable "name" {
  description = "Name of the Config aggregator"
  type        = string
  default     = "org-config-aggregator"
}

variable "aggregation_role_arn" {
  description = "IAM role ARN that AWS Config assumes to aggregate data"
  type        = string
}

variable "all_regions" {
  description = "Whether to aggregate from all regions"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags to apply to the Config aggregator"
  type        = map(string)
  default     = {}
}
