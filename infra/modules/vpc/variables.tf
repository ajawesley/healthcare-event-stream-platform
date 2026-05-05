variable "name" {
  type = string
}

variable "region" {
  type = string
}

variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

variable "primary_az" {
  type = string
}

variable "azs" {
  type = map(object({
    index = number
  }))
}

variable "tags" {
  type    = map(string)
  default = {}
}
