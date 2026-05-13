############################################
# Inspector ORG-level Delegated Administrator
############################################

resource "aws_inspector2_organization_admin_account" "this" {
  admin_account_id = var.admin_account_id
}

############################################
# Inspector ORG-level Configuration
############################################

resource "aws_inspector2_organization_configuration" "this" {
  auto_enable {
    ec2_scan_enabled     = true
    ecr_scan_enabled     = true
    lambda_scan_enabled  = true
  }

  depends_on = [
    aws_inspector2_organization_admin_account.this
  ]
}
