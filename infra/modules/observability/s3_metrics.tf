############################################
# S3 Object-Level Metric Filters
############################################

# Count of ingested objects (PutObject / CompleteMultipartUpload)
resource "aws_cloudwatch_log_metric_filter" "s3_ingested_objects" {
  name           = "hesp-${var.environment}-s3-ingested-objects"
  log_group_name = aws_cloudwatch_log_group.s3_data_events.name

  pattern = "{ ($.eventName = \"PutObject\" || $.eventName = \"CompleteMultipartUpload\") }"

  metric_transformation {
    name      = "s3_ingested_objects"
    namespace = "HESP"
    value     = "1"
  }
}

# Total bytes ingested (per event)
resource "aws_cloudwatch_log_metric_filter" "s3_ingested_bytes" {
  name           = "hesp-${var.environment}-s3-ingested-bytes"
  log_group_name = aws_cloudwatch_log_group.s3_data_events.name

  pattern = "{ ($.eventName = \"PutObject\" || $.eventName = \"CompleteMultipartUpload\") && $.additionalEventData.bytesTransferredIn > 0 }"

  metric_transformation {
    name      = "s3_ingested_bytes"
    namespace = "HESP"
    value     = "$.additionalEventData.bytesTransferredIn"
  }
}

# Large objects (e.g., > 5MB) - potential anomaly for HL7
resource "aws_cloudwatch_log_metric_filter" "s3_large_objects" {
  name           = "hesp-${var.environment}-s3-large-objects"
  log_group_name = aws_cloudwatch_log_group.s3_data_events.name

  pattern = "{ ($.eventName = \"PutObject\" || $.eventName = \"CompleteMultipartUpload\") && $.additionalEventData.bytesTransferredIn > 5242880 }"

  metric_transformation {
    name      = "s3_large_objects"
    namespace = "HESP"
    value     = "1"
  }
}

# Zero-byte objects - likely ingestion corruption
resource "aws_cloudwatch_log_metric_filter" "s3_zero_byte_objects" {
  name           = "hesp-${var.environment}-s3-zero-byte-objects"
  log_group_name = aws_cloudwatch_log_group.s3_data_events.name

  pattern = "{ ($.eventName = \"PutObject\" || $.eventName = \"CompleteMultipartUpload\") && $.additionalEventData.bytesTransferredIn = 0 }"

  metric_transformation {
    name      = "s3_zero_byte_objects"
    namespace = "HESP"
    value     = "1"
  }
}
