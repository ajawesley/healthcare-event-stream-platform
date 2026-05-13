terraform {
  required_version = ">= 1.5.0"
}

# ---------------------------------------------------------
# Import permission boundary from landing-zone/iam-boundaries
# ---------------------------------------------------------
module "iam_boundaries" {
  source = "../../landing-zone/iam-boundaries"
}

# ---------------------------------------------------------
# Glue Assume Role Policy
# ---------------------------------------------------------
data "aws_iam_policy_document" "glue_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["glue.amazonaws.com"]
    }
  }
}

# ---------------------------------------------------------
# Glue Job Role
# ---------------------------------------------------------
resource "aws_iam_role" "glue_role" {
  name               = "hesp-glue-job-role"
  assume_role_policy = data.aws_iam_policy_document.glue_assume.json

  # Enforced by SCP — must be present
  permissions_boundary = module.iam_boundaries.permission_boundary_arn

  tags = var.tags
}

# ---------------------------------------------------------
# Attach AWS-managed Glue policies
# ---------------------------------------------------------
resource "aws_iam_role_policy_attachment" "glue_service" {
  role       = aws_iam_role.glue_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSGlueServiceRole"
}

resource "aws_iam_role_policy_attachment" "glue_s3_access" {
  role       = aws_iam_role.glue_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonS3FullAccess"
}

# ---------------------------------------------------------
# Outputs
# ---------------------------------------------------------
output "glue_role_arn" {
  value = aws_iam_role.glue_role.arn
}
