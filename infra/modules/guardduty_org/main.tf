############################################
# GuardDuty Detector (ORG Management Account)
############################################

resource "aws_guardduty_detector" "org" {
  enable = true

  tags = merge(
    var.tags,
    { Name = "guardduty-org-detector-${var.region}" }
  )
}

############################################
# Designate Delegated Administrator
############################################

resource "aws_guardduty_organization_admin_account" "this" {
  admin_account_id = var.admin_account_id

  depends_on = [aws_guardduty_detector.org]
}

############################################
# Organization Configuration (auto-enroll)
############################################

resource "aws_guardduty_organization_configuration" "this" {
  detector_id = aws_guardduty_detector.org.id

  auto_enable_organization_members = "ALL"

  depends_on = [aws_guardduty_organization_admin_account.this]
}

############################################
# Add Existing Accounts as Members
############################################

resource "aws_guardduty_member" "members" {
  for_each = toset(var.member_account_ids)

  account_id  = each.value
  detector_id = aws_guardduty_detector.org.id
  email       = "noreply@example.com"

  invite                     = true
  disable_email_notification = true

  depends_on = [aws_guardduty_organization_admin_account.this]

  tags = merge(
    var.tags,
    { Name = "guardduty-member-${each.value}" }
  )
}
