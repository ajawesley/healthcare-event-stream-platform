############################################
# Enable Inspector2 in Member Account
############################################

resource "aws_inspector2_account" "this" {
  # No arguments required — enables Inspector2 locally
}

############################################
# Accept Inspector Delegated Admin
############################################

resource "aws_inspector2_organization_admin_account_association" "this" {
  admin_account_id = var.admin_account_id

  depends_on = [
    aws_inspector2_account.this
  ]
}

############################################
# Enable Inspector2 Scanning Features
############################################

resource "aws_inspector2_enabler" "this" {
  account_ids = [var.member_account_id]

  resource_types = [
    "EC2",
    "ECR",
    "LAMBDA"
  ]

  depends_on = [
    aws_inspector2_account.this
  ]
}
