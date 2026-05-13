############################################
# Secondary AWS provider (aliased)
############################################

provider "aws" {
  alias  = "default_region"
  region = var.aws_region
}
