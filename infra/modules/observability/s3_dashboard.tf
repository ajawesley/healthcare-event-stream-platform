resource "aws_cloudwatch_dashboard" "s3_ingestion" {
  dashboard_name = "hesp-${var.environment}-s3-ingestion"

  dashboard_body = jsonencode({
    widgets = [
      {
        "type" : "text",
        "x" : 0,
        "y" : 0,
        "width" : 24,
        "height" : 2,
        "properties" : {
          "markdown" : "# HESP ${var.environment} – S3 Ingestion Metrics\nObject count, size, and anomalies."
        }
      },
      {
        "type" : "metric",
        "x" : 0,
        "y" : 2,
        "width" : 12,
        "height" : 6,
        "properties" : {
          "title" : "S3 Objects Ingested per Hour",
          "view" : "timeSeries",
          "stacked" : false,
          "metrics" : [
            ["HESP", "s3_ingested_objects", "Environment", var.environment]
          ],
          "region" : var.aws_region,
          "stat" : "Sum",
          "period" : 3600
        }
      },
      {
        "type" : "metric",
        "x" : 12,
        "y" : 2,
        "width" : 12,
        "height" : 6,
        "properties" : {
          "title" : "S3 Bytes Ingested per Hour",
          "view" : "timeSeries",
          "stacked" : false,
          "metrics" : [
            ["HESP", "s3_ingested_bytes", "Environment", var.environment]
          ],
          "region" : var.aws_region,
          "stat" : "Sum",
          "period" : 3600
        }
      },
      {
        "type" : "metric",
        "x" : 0,
        "y" : 8,
        "width" : 12,
        "height" : 6,
        "properties" : {
          "title" : "Large Objects (>5MB)",
          "view" : "timeSeries",
          "stacked" : false,
          "metrics" : [
            ["HESP", "s3_large_objects", "Environment", var.environment]
          ],
          "region" : var.aws_region,
          "stat" : "Sum",
          "period" : 3600
        }
      },
      {
        "type" : "metric",
        "x" : 12,
        "y" : 8,
        "width" : 12,
        "height" : 6,
        "properties" : {
          "title" : "Zero-Byte Objects",
          "view" : "timeSeries",
          "stacked" : false,
          "metrics" : [
            ["HESP", "s3_zero_byte_objects", "Environment", var.environment]
          ],
          "region" : var.aws_region,
          "stat" : "Sum",
          "period" : 3600
        }
      }
    ]
  })
}
