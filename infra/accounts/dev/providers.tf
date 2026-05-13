provider "aws" {
  region = var.aws_region

  assume_role {
    role_arn = var.workload_account_role_arn
  }
}
