resource "aws_glue_catalog_database" "acmecorp_hesp" {
  name = "acmecorp_hesp_db"

  tags = var.tags
}
