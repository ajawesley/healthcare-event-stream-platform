variable "aws_region" {
  type = string
}

variable "org_name" {
  type = string
}

variable "owner" {
  type = string
}

variable "extra_tags" {
  type    = map(string)
  default = {}
}

variable "github_owner" {
  type = string
}

variable "github_repo" {
  type = string
}
