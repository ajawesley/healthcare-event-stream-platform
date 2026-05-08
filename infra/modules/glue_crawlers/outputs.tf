output "crawler_role_arn" {
  value = aws_iam_role.crawler_role.arn
}

output "events_crawler_name" {
  value = aws_glue_crawler.events.name
}

output "errors_crawler_name" {
  value = aws_glue_crawler.errors.name
}

output "database_name" {
  value = aws_glue_catalog_database.acmecorp_hesp.name
}
