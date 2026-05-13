############################################
# GuardDuty Member
############################################

resource "aws_guardduty_detector" "detector" {
  enable = true
}

resource "aws_guardduty_member" "member" {
  account_id  = var.security_admin_account_id
  detector_id = aws_guardduty_detector.detector.id
  email       = "security@example.com"
}

############################################
# Security Hub
############################################

resource "aws_securityhub_account" "this" {}

resource "aws_securityhub_organization_admin_account_association" "assoc" {
  admin_account_id = var.security_admin_account_id
}

############################################
# Inspector2
############################################

resource "aws_inspector2_account" "this" {}

resource "aws_inspector2_enabler" "enable" {
  account_ids    = [var.security_admin_account_id]
  resource_types = ["EC2", "ECR", "LAMBDA"]
}
