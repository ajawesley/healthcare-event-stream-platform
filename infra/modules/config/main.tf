############################################
# AWS Config Recorder
############################################

resource "aws_config_configuration_recorder" "this" {
  name     = "${var.name_prefix}-config-recorder"
  role_arn = var.config_role_arn

  recording_group {
    all_supported                 = true
    include_global_resource_types = true
  }
}

############################################
# AWS Config Delivery Channel
############################################

resource "aws_config_delivery_channel" "this" {
  name           = "${var.name_prefix}-config-delivery"
  s3_bucket_name = var.log_archive_bucket_name

  # IMPORTANT:
  # Removed s3_kms_key_arn because the log archive bucket uses AES256 SSE.
  # Leaving this set causes AWS to reject the delivery channel.
  #
  # s3_kms_key_arn = var.kms_key_arn

  snapshot_delivery_properties {
    delivery_frequency = "TwentyFour_Hours"
  }

  depends_on = [
    aws_config_configuration_recorder.this
  ]
}

############################################
# AWS Config Recorder Status
############################################

resource "aws_config_configuration_recorder_status" "this" {
  name       = aws_config_configuration_recorder.this.name
  is_enabled = true

  depends_on = [
    aws_config_delivery_channel.this
  ]
}

############################################
# OPTIONAL: Config Rules (scaffolding)
############################################

# resource "aws_config_config_rule" "required_tags" {
#   name = "${var.name_prefix}-required-tags"
#
#   source {
#     owner             = "AWS"
#     source_identifier = "REQUIRED_TAGS"
#   }
#
#   input_parameters = jsonencode({
#     tag1Key = "Owner"
#     tag2Key = "Environment"
#   })
#
#   depends_on = [
#     aws_config_configuration_recorder_status.this
#   ]
#
#   tags = var.tags
# }
