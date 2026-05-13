terraform {
  required_version = ">= 1.5.0"
}

# Import the permission boundary from the landing-zone/iam-boundaries module
module "iam_boundaries" {
  source = "../../landing-zone/iam-boundaries"
}

# You can also pass this in via a variable if accounts are split by workspace
