############################################
# Glue Database for Centralized Logs
############################################

resource "aws_glue_catalog_database" "logs" {
  name = "${var.name_prefix}_logs"
}

############################################
# Glue Table for CloudWatch Logs
############################################

resource "aws_glue_catalog_table" "cw_logs" {
  name          = "cloudwatch_logs"
  database_name = aws_glue_catalog_database.logs.name

  table_type = "EXTERNAL_TABLE"

  storage_descriptor {
    location      = "s3://${var.log_archive_bucket_name}/cloudwatch/"
    input_format  = "org.apache.hadoop.mapred.TextInputFormat"
    output_format = "org.apache.hadoop.hive.ql.io.HiveIgnoreKeyTextOutputFormat"

    serde_info {
      serialization_library = "org.openx.data.jsonserde.JsonSerDe"
    }

    columns = [
      {
        name = "message"
        type = "string"
      },
      {
        name = "timestamp"
        type = "bigint"
      }
    ]
  }
}

############################################
# Athena Workgroup
############################################

resource "aws_athena_workgroup" "logs" {
  name = "${var.name_prefix}-logs-wg"

  configuration {
    enforce_workgroup_configuration = true

    result_configuration {
      output_location = "s3://${var.log_archive_bucket_name}/athena-results/"
    }
  }

  tags = var.tags
}
