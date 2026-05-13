variable "role_name" {
  description = "Name of the IAM role GitHub Actions will assume."
  type        = string
}

variable "github_owner" {
  description = "GitHub username or organization name."
  type        = string
}

variable "github_repo" {
  description = "GitHub repository name."
  type        = string
}

variable "github_ref" {
  description = "GitHub ref pattern allowed to assume the role (e.g., 'refs/heads/main' or '*')."
  type        = string
  default     = "*"
}

variable "inline_policy_statements" {
  description = "List of IAM policy statements to attach to the role."
  type        = list(any)
}
