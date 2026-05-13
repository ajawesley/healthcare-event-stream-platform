############################################
# Enable Security Hub in Delegated Admin Account
############################################

resource "aws_securityhub_account" "this" {
  # No arguments required — enables Security Hub in this account
}

############################################
# Enable All Security Hub Standards
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
