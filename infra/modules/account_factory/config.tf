############################################
# AWS Config Recorder
############################################

resource "aws_config_configuration_recorder" "recorder" {
  name     = "default"
  role_arn = aws_iam_role.config.arn
}

############################################
# IAM Role for Config
############################################

resource "aws_iam_role" "config" {
  name = "AWSConfigRole"

  assume_role_policy = data.aws_iam_policy_document.config_assume.json
}

data "aws_iam_policy_document" "config_assume" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["config.amazonaws.com"]
    }
    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role_policy_attachment" "config_attach" {
  role       = aws_iam_role.config.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSConfigRole"
}

############################################
# Delivery Channel
############################################

resource "aws_config_delivery_channel" "channel" {
  name           = "default"
  s3_bucket_name = var.config_bucket_name

  depends_on = [aws_config_configuration_recorder.recorder]
}
