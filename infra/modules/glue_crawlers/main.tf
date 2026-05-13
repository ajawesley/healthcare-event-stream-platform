resource "aws_glue_crawler" "events" {
  name          = "acmecorp-hesp-events-crawler-${var.environment}"
  role          = aws_iam_role.crawler_role.arn
  database_name = aws_glue_catalog_database.acmecorp_hesp.name

  s3_target {
    path = "s3://${var.events_bucket}/"
  }

  configuration = jsonencode({
    Version = 1.0
    CrawlerOutput = {
      Partitions = { AddOrUpdateBehavior = "InheritFromTable" }
    }
  })

  schema_change_policy {
    update_behavior = "UPDATE_IN_DATABASE"
    delete_behavior = "LOG"
  }

  schedule = "cron(0/5 * * * ? *)"

  tags = var.tags
}

resource "aws_glue_crawler" "errors" {
  name          = "acmecorp-hesp-errors-crawler-${var.environment}"
  role          = aws_iam_role.crawler_role.arn
  database_name = aws_glue_catalog_database.acmecorp_hesp.name

  s3_target {
    path = "s3://${var.errors_bucket}/"
  }

  configuration = jsonencode({
    Version = 1.0
    CrawlerOutput = {
      Partitions = { AddOrUpdateBehavior = "InheritFromTable" }
    }
  })

  schema_change_policy {
    update_behavior = "UPDATE_IN_DATABASE"
    delete_behavior = "LOG"
  }

  schedule = "cron(0/15 * * * ? *)"

  tags = var.tags
}
