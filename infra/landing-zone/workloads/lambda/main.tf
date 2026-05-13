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
# Lambda Assume Role Policy
# ---------------------------------------------------------
data "aws_iam_policy_document" "lambda_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

# ---------------------------------------------------------
# Lambda Execution Role
# ---------------------------------------------------------
resource "aws_iam_role" "lambda_exec_role" {
  name               = "hesp-lambda-exec-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json

  # Enforced by SCP — must be present
  permissions_boundary = module.iam_boundaries.permission_boundary_arn

  tags = var.tags
}

# ---------------------------------------------------------
# Attach AWS-managed Lambda policies
# ---------------------------------------------------------
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "lambda_vpc_access" {
  role       = aws_iam_role.lambda_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

# ---------------------------------------------------------
# Outputs
# ---------------------------------------------------------
output "lambda_exec_role_arn" {
  value = aws_iam_role.lambda_exec_role.arn
}
