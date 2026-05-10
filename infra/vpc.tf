############################################
# VPC Module
############################################

module "vpc" {
  source = "./modules/vpc"

  name       = "${var.app_name}-${var.environment}"
  region     = var.aws_region
  vpc_cidr   = "10.0.0.0/16"
  primary_az = "us-east-1a"

  azs = {
    "us-east-1a" = { index = 0 }
    "us-east-1b" = { index = 1 }
    "us-east-1c" = { index = 2 }
  }

  tags = local.base_tags
}
