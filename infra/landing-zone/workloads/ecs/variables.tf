variable "tags" {
  type        = map(string)
  default     = {}
  description = "Common tags for ECS IAM resources"
}
