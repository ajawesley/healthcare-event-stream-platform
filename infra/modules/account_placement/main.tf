############################################
# Move an existing AWS account into an OU
############################################

resource "aws_organizations_account" "placement" {
  account_id = var.account_id
  name       = var.account_name
  email      = var.account_email

  parent_id = var.ou_id

  # Tags are optional but recommended
  tags = var.tags
}
