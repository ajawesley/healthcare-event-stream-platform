terraform {
  required_version = ">= 1.5.0"
}

provider "aws" {
  region = var.aws_region
}

locals {
  tags = merge(
    {
      Project = var.org_name
      Owner   = var.owner
    },
    var.extra_tags
  )
}

############################################
# Permission Boundaries
############################################

module "permission_boundaries" {
  source = "./permission-boundaries"

}

############################################
# OIDC Provider + Roles
############################################

module "oidc" {
  source = "./oidc"

}

############################################
# SCPs (optional)
############################################

module "scp" {
  source = "./scp"

}
