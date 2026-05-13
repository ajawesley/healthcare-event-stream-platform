variable "org_name" {
  description = "Logical name for the organization / project"
  type        = string
}

variable "github_owner" {
  description = "GitHub organization or user that owns the repo"
  type        = string
}

variable "github_repo" {
  description = "GitHub repository name used for deployments"
  type        = string
}

variable "owner" {
  description = "Owner tag applied to all resources"
  type        = string
}

variable "tags" {
  description = "Common tags applied to org-level resources"
  type        = map(string)
}
