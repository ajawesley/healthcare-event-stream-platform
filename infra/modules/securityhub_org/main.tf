############################################
# Enable Security Hub in ORG Management Account
############################################

resource "aws_securityhub_organization_admin_account" "this" {
  admin_account_id = var.admin_account_id
}

############################################
# Enable Security Hub Organization Configuration
############################################

resource "aws_securityhub_organization_configuration" "this" {
  auto_enable = true

  depends_on = [
    aws_securityhub_organization_admin_account.this
  ]
}
