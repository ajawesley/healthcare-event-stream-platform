############################################
# SNS Topic for Alerts
############################################

resource "aws_sns_topic" "hesp_alerts" {
  name = "hesp-${var.environment}-alerts"
}

############################################
# 1. Glue Job Failure Alarm
############################################

resource "aws_cloudwatch_metric_alarm" "glue_job_failures" {
  alarm_name          = "hesp-${var.environment}-glue-job-failures"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "glue_job_failures"
  namespace           = "HESP"
  period              = 300
  statistic           = "Sum"
  threshold           = 0
  alarm_description   = "Triggers when Glue job reports any failures."
  alarm_actions       = [aws_sns_topic.hesp_alerts.arn]
}

############################################
# 2. Invalid Records Alarm
############################################

resource "aws_cloudwatch_metric_alarm" "invalid_records" {
  alarm_name          = "hesp-${var.environment}-invalid-records"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "invalid_records"
  namespace           = "HESP/Glue"
  period              = 300
  statistic           = "Sum"
  threshold           = 100
  alarm_description   = "Triggers when invalid records exceed threshold."
  alarm_actions       = [aws_sns_topic.hesp_alerts.arn]
}

############################################
# 3. Schema Drift Alarm
############################################

resource "aws_cloudwatch_metric_alarm" "schema_drift" {
  alarm_name          = "hesp-${var.environment}-schema-drift"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "glue_schema_drift_events"
  namespace           = "HESP"
  period              = 300
  statistic           = "Sum"
  threshold           = 0
  alarm_description   = "Triggers when schema drift is detected in Glue job."
  alarm_actions       = [aws_sns_topic.hesp_alerts.arn]
}

############################################
# 4. S3 Ingestion Anomaly Alarm (Volume Drop)
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
  alarm_description   = "Triggers when ingestion volume drops unexpectedly."
  alarm_actions       = [aws_sns_topic.hesp_alerts.arn]
}

############################################
# 5. S3 Ingestion Spike Alarm (Volume Spike)
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
  alarm_description   = "Triggers when ingestion volume spikes unexpectedly."
  alarm_actions       = [aws_sns_topic.hesp_alerts.arn]
}

############################################
# 6. Pipeline Latency Alarm (End-to-End)
############################################

resource "aws_cloudwatch_metric_alarm" "pipeline_latency" {
  alarm_name          = "hesp-${var.environment}-pipeline-latency"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "pipeline_latency_ms_end_to_end"
  namespace           = "HESP"
  period              = 300
  statistic           = "Average"
  threshold           = 300000   # 5 minutes
  alarm_description   = "Triggers when end-to-end pipeline latency exceeds 5 minutes."
  alarm_actions       = [aws_sns_topic.hesp_alerts.arn]
}

############################################
# 7. Replay Spike Alarm
############################################

resource "aws_cloudwatch_metric_alarm" "replay_spike" {
  alarm_name          = "hesp-${var.environment}-replay-spike"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "replay_events"
  namespace           = "HESP"
  period              = 300
  statistic           = "Sum"
  threshold           = 10
  alarm_description   = "Triggers when replayed events exceed threshold."
  alarm_actions       = [aws_sns_topic.hesp_alerts.arn]
}

############################################
# 8. Late Arrival Spike Alarm
############################################

resource "aws_cloudwatch_metric_alarm" "late_arrival_spike" {
  alarm_name          = "hesp-${var.environment}-late-arrival-spike"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "late_arrival_events"
  namespace           = "HESP"
  period              = 300
  statistic           = "Sum"
  threshold           = 10
  alarm_description   = "Triggers when late-arriving events exceed threshold."
  alarm_actions       = [aws_sns_topic.hesp_alerts.arn]
}
