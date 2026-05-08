resource "aws_iam_role" "crawler_role" {
  name = "acmecorp-hesp-crawler-role-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = { Service = "glue.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })

  tags = var.tags
}

resource "aws_iam_role_policy" "crawler_policy" {
  name = "acmecorp-hesp-crawler-policy-${var.environment}"
  role = aws_iam_role.crawler_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws:s3:::${var.events_bucket}",
          "arn:aws:s3:::${var.events_bucket}/*",
          "arn:aws:s3:::${var.errors_bucket}",
          "arn:aws:s3:::${var.errors_bucket}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "glue:CreateTable",
          "glue:UpdateTable",
          "glue:GetTable",
          "glue:GetTables",
          "glue:GetDatabase",
          "glue:GetDatabases"
        ]
        Resource = "*"
      }
    ]
  })
}
