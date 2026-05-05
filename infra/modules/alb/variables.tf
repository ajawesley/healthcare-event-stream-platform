variable "app_name" {
  type = string
}

variable "environment" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "subnet_ids" {
  type = list(string)
}

variable "owner" {
  type = string
}

variable "cost_center" {
  type = string
}

variable "tags" {
  type = map(string)
}

variable "alb_security_group_id" {
  type = string
}

