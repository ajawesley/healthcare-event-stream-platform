variable "app_name" {}
variable "environment" {}
variable "aws_region" {}

variable "owner" {}
variable "cost_center" {}

variable "raw_bucket_name" {}
variable "access_logs_bucket_name" {}

variable "container_image" {}

variable "desired_count" {
  type    = number
  default = 1
}

variable "compliance_db_password_secret_arn" {}
variable "compliance_db_username" {}
variable "compliance_db_password" {}
variable "compliance_db_name" {}

variable "glue_script_s3_path" {}
variable "glue_temp_dir" {}

variable "s3_output_base_path" {}
variable "s3_error_path" {}

variable "lambda_zip_path" {}
variable "adot_image" {}
