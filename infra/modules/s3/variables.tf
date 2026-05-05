variable "bucket_name" {
  type        = string
  description = "Name of the raw events bucket."
}

variable "environment" {
  type        = string
  description = "Deployment environment."
}

variable "owner" {
  type        = string
  description = "Owner tag."
}

variable "cost_center" {
  type        = string
  description = "Cost center tag."
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "Additional tags."
}

