output "athena_workgroup" {
  value = aws_athena_workgroup.logs.name
}

output "glue_database" {
  value = aws_glue_catalog_database.logs.name
}

output "glue_table" {
  value = aws_glue_catalog_table.cw_logs.name
}
