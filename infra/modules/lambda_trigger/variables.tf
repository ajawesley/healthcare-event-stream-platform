variable "app_name" {
  type = string
}

variable "environment" {
  type = string
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

variable "glue_job_name" {
  type = string
}

variable "glue_job_arn" {
  type = string
}

variable "raw_bucket_name" {
  type = string
}

variable "lambda_role_arn" {
  type = string
}

variable "lambda_role_name" {
  type = string
}

variable "lambda_zip_path" {
  type = string
}
