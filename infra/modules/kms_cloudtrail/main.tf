############################################
# KMS Key for CloudTrail (ORG-level)
############################################

resource "aws_kms_key" "this" {
  description         = "KMS key for CloudTrail logs"
  enable_key_rotation = true

  tags = merge(
    var.tags,
    { Name = "${var.name_prefix}-cloudtrail-kms" }
  )
}

resource "aws_kms_alias" "this" {
  name          = "alias/${var.name_prefix}-cloudtrail"
  target_key_id = aws_kms_key.this.key_id
}
