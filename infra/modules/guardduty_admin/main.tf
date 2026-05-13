resource "aws_guardduty_detector" "admin" {
  enable = true

  tags = merge(
    var.tags,
    { Name = "${var.name_prefix}-guardduty-admin-detector" }
  )
}
