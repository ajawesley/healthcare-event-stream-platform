############################################
# SNS Topic for S3 Alerts (reuse if exists)
############################################

resource "aws_sns_topic" "hesp_s3_alerts" {
  name = "hesp-${var.environment}-s3-alerts"
}

############################################
# 1. Ingestion Volume Drop (Objects / Hour)
############################################

resource "aws_cloudwatch_metric_alarm" "s3_ingestion_drop" {
  alarm_name          = "hesp-${var.environment}-s3-ingestion-drop"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 1
  metric_name         = "s3_ingested_objects"
  namespace           = "HESP"
  period              = 3600
  statistic           = "Sum"
  threshold           = 10
  alarm_description   = "Triggers when S3 ingestion volume drops below expected threshold."
  alarm_actions       = [aws_sns_topic.hesp_s3_alerts.arn]
}

############################################
# 2. Ingestion Volume Spike (Objects / Hour)
############################################

resource "aws_cloudwatch_metric_alarm" "s3_ingestion_spike" {
  alarm_name          = "hesp-${var.environment}-s3-ingestion-spike"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "s3_ingested_objects"
  namespace           = "HESP"
  period              = 3600
  statistic           = "Sum"
  threshold           = 50000
  alarm_description   = "Triggers when S3 ingestion volume spikes unexpectedly (possible replay storm)."
  alarm_actions       = [aws_sns_topic.hesp_s3_alerts.arn]
}

############################################
# 3. Ingestion Size Spike (Bytes / Hour)
############################################

resource "aws_cloudwatch_metric_alarm" "s3_ingestion_size_spike" {
  alarm_name          = "hesp-${var.environment}-s3-ingestion-size-spike"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "s3_ingested_bytes"
  namespace           = "HESP"
  period              = 3600
  statistic           = "Sum"
  threshold           = 10737418240 # 10 GB
  alarm_description   = "Triggers when S3 ingestion bytes per hour spike unexpectedly."
  alarm_actions       = [aws_sns_topic.hesp_s3_alerts.arn]
}

############################################
# 4. Large Object Spike
############################################

resource "aws_cloudwatch_metric_alarm" "s3_large_objects" {
  alarm_name          = "hesp-${var.environment}-s3-large-objects"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "s3_large_objects"
  namespace           = "HESP"
  period              = 3600
  statistic           = "Sum"
  threshold           = 10
  alarm_description   = "Triggers when large S3 objects (>5MB) exceed threshold."
  alarm_actions       = [aws_sns_topic.hesp_s3_alerts.arn]
}

############################################
# 5. Zero-Byte Objects
############################################

resource "aws_cloudwatch_metric_alarm" "s3_zero_byte_objects" {
  alarm_name          = "hesp-${var.environment}-s3-zero-byte-objects"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "s3_zero_byte_objects"
  namespace           = "HESP"
  period              = 3600
  statistic           = "Sum"
  threshold           = 0
  alarm_description   = "Triggers when zero-byte S3 objects are detected (likely ingestion corruption)."
  alarm_actions       = [aws_sns_topic.hesp_s3_alerts.arn]
}
