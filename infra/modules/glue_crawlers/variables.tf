variable "environment" {
  type = string
}

variable "tags" {
  type = map(string)
}

variable "events_bucket" {
  type = string
}

variable "errors_bucket" {
  type = string
}
