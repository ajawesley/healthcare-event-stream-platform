# Glue job failure metric from logs
resource "aws_cloudwatch_log_metric_filter" "glue_job_failed" {
  name           = "hesp-${var.environment}-glue-job-failed"
  log_group_name = "/aws/glue/${var.app_name}-${var.environment}"

  pattern = "{ ($.message = \"job_failed\" || $.event = \"job_failed\") }"

  metric_transformation {
    name      = "glue_job_failures"
    namespace = "HESP"
    value     = "1"
  }
}

# Invalid records metric
#resource "aws_cloudwatch_log_metric_filter" "glue_invalid_records" {
#  name           = "hesp-${var.environment}-glue-invalid-records"
#  log_group_name = "/aws/glue/${var.app_name}-${var.environment}"
#
#  pattern = "{ $.message = \"validation_results\" && $.invalid_records > 0 }"
#
#  metric_transformation {
#    name      = "glue_invalid_records"
#    namespace = "HESP"
#    value     = "$.invalid_records"
#  }
#}
#
# Schema drift metric (assuming you log 'schema_drift_detected')
#resource "aws_cloudwatch_log_metric_filter" "glue_schema_drift" {
#  name           = "hesp-${var.environment}-glue-schema-drift"
#  log_group_name = "/aws/glue/${var.app_name}-${var.environment}"
#
#  pattern = "{ $.message = \"schema_drift_detected\" }"
#
#  metric_transformation {
#    name      = "glue_schema_drift_events"
#    namespace = "HESP"
#    value     = "1"
#  }
#}
