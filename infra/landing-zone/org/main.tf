terraform {
  required_version = ">= 1.5.0"
}

provider "aws" {
  region = var.aws_region
}

############################################
# Locals
############################################

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
# OIDC Provider + GitHub Deploy Role
############################################

module "oidc" {
  source       = "./oidc"
  owner        = var.owner
  org_name     = var.org_name
  github_owner = var.github_owner
  github_repo  = var.github_repo
  tags         = local.tags
}

############################################
# SCPs (optional)
############################################

module "scp" {
  source = "./scp"

}
