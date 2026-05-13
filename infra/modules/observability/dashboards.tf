resource "aws_cloudwatch_dashboard" "pipeline_observability" {
  dashboard_name = "hesp-${var.environment}-pipeline-observability"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "text"
        x      = 0
        y      = 0
        width  = 24
        height = 2
        properties = {
          markdown = "# HESP ${var.environment} – Pipeline Observability\nLatency, errors, lineage health, and S3 ingestion metrics."
        }
      },

      # -------------------------------------------------------------------
      # Pipeline Latency (Phase 2.5)
      # -------------------------------------------------------------------
      {
        type   = "metric"
        x      = 0
        y      = 2
        width  = 12
        height = 6
        properties = {
          title   = "Ingest → Glue End-to-End Latency (p50/p95/p99)"
          view    = "timeSeries"
          stacked = false
          metrics = [
            ["HESP", "pipeline_latency_ms_end_to_end_p50", "Environment", var.environment],
            [".", "pipeline_latency_ms_end_to_end_p95", "Environment", var.environment],
            [".", "pipeline_latency_ms_end_to_end_p99", "Environment", var.environment]
          ]
          region = var.aws_region
          stat   = "Average"
          period = 60
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 2
        width  = 12
        height = 6
        properties = {
          title   = "Stage Latency (Ingest→Canon, Canon→Write, Write→S3, S3→Glue)"
          view    = "timeSeries"
          stacked = false
          metrics = [
            ["HESP", "pipeline_latency_ms_ingest_to_canonical", "Environment", var.environment],
            [".", "pipeline_latency_ms_canonical_to_write", "Environment", var.environment],
            [".", "pipeline_latency_ms_write_to_s3", "Environment", var.environment],
            [".", "pipeline_latency_ms_s3_to_glue", "Environment", var.environment]
          ]
          region = var.aws_region
          stat   = "Average"
          period = 60
        }
      },

      # -------------------------------------------------------------------
      # Glue Errors / Schema Drift (Phase 2.5)
      # -------------------------------------------------------------------
      {
        type   = "metric"
        x      = 0
        y      = 8
        width  = 12
        height = 6
        properties = {
          title   = "Glue Job Failures"
          view    = "timeSeries"
          stacked = false
          metrics = [
            ["HESP", "glue_job_failures", "Environment", var.environment]
          ]
          region = var.aws_region
          stat   = "Sum"
          period = 300
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 8
        width  = 12
        height = 6
        properties = {
          title   = "Invalid Records / Schema Errors"
          view    = "timeSeries"
          stacked = false
          metrics = [
            ["HESP", "glue_invalid_records", "Environment", var.environment],
            [".", "glue_schema_drift_events", "Environment", var.environment]
          ]
          region = var.aws_region
          stat   = "Sum"
          period = 300
        }
      },

      # -------------------------------------------------------------------
      # S3 Ingestion Metrics (Phase 2.6)
      # -------------------------------------------------------------------
      {
        type   = "metric"
        x      = 0
        y      = 14
        width  = 12
        height = 6
        properties = {
          title   = "S3 Ingestion Volume (Objects / Hour)"
          view    = "timeSeries"
          stacked = false
          metrics = [
            ["HESP", "s3_ingested_objects", "Environment", var.environment]
          ]
          region = var.aws_region
          stat   = "Sum"
          period = 3600
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 14
        width  = 12
        height = 6
        properties = {
          title   = "S3 Ingestion Size (Bytes / Hour)"
          view    = "timeSeries"
          stacked = false
          metrics = [
            ["HESP", "s3_ingested_bytes", "Environment", var.environment]
          ]
          region = var.aws_region
          stat   = "Sum"
          period = 3600
        }
      },

      # -------------------------------------------------------------------
      # S3 Anomaly Metrics (Phase 2.6)
      # -------------------------------------------------------------------
      {
        type   = "metric"
        x      = 0
        y      = 20
        width  = 12
        height = 6
        properties = {
          title   = "Large S3 Objects (>5MB)"
          view    = "timeSeries"
          stacked = false
          metrics = [
            ["HESP", "s3_large_objects", "Environment", var.environment]
          ]
          region = var.aws_region
          stat   = "Sum"
          period = 3600
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 20
        width  = 12
        height = 6
        properties = {
          title   = "Zero-Byte S3 Objects"
          view    = "timeSeries"
          stacked = false
          metrics = [
            ["HESP", "s3_zero_byte_objects", "Environment", var.environment]
          ]
          region = var.aws_region
          stat   = "Sum"
          period = 3600
        }
      }
    ]
  })
}
