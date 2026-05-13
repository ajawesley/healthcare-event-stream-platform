############################################
# Enable Security Hub in Member Account
############################################

resource "aws_securityhub_account" "this" {
  # No arguments required — enables Security Hub locally
}

############################################
# Accept Security Hub Delegated Admin
############################################

resource "aws_securityhub_organization_admin_account_association" "this" {
  admin_account_id = var.admin_account_id

  depends_on = [
    aws_securityhub_account.this
  ]
}

############################################
# Enable Security Hub Standards (Member Account)
############################################

resource "aws_securityhub_standards_subscription" "aws_foundational" {
  standards_arn = "arn:aws:securityhub:::ruleset/aws-foundational-security-best-practices/v/1.0.0"

  depends_on = [
    aws_securityhub_account.this
  ]
}

resource "aws_securityhub_standards_subscription" "cis" {
  standards_arn = "arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"

  depends_on = [
    aws_securityhub_account.this
  ]
}
