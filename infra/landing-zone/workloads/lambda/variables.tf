variable "tags" {
  type        = map(string)
  default     = {}
  description = "Common tags for Lambda IAM resources"
}
