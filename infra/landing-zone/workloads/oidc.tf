############################################
# GitHub OIDC Role for Landing Zone Phase 2 (Workloads)
############################################

resource "aws_iam_role" "github_oidc_workloads" {
  name = "github-oidc-workloads"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = "arn:aws:iam::<ORG_ACCOUNT_ID>:oidc-provider/token.actions.githubusercontent.com"
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          }
          StringLike = {
            "token.actions.githubusercontent.com:sub" = "repo:ajawesley/healthcare-event-stream-platform:*"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "github_oidc_workloads_policy" {
  role = aws_iam_role.github_oidc_workloads.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ec2:*",
          "iam:*",
          "s3:*",
          "kms:*",
          "cloudtrail:*",
          "config:*",
          "guardduty:*",
          "securityhub:*",
          "logs:*"
        ]
        Resource = "*"
      }
    ]
  })
}

############################################
# GitHub OIDC Roles for Application Environments
############################################

locals {
  environments = ["dev", "qa", "prod"]
}

resource "aws_iam_role" "github_oidc_env" {
  for_each = toset(local.environments)

  name = "github-oidc-${each.key}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = "arn:aws:iam::<ORG_ACCOUNT_ID>:oidc-provider/token.actions.githubusercontent.com"
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          }
          StringLike = {
            "token.actions.githubusercontent.com:sub" = "repo:ajawesley/healthcare-event-stream-platform:*"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "github_oidc_env_policy" {
  for_each = aws_iam_role.github_oidc_env

  role = each.value.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecs:*",
          "ecr:*",
          "lambda:*",
          "glue:*",
          "s3:*",
          "iam:PassRole",
          "cloudwatch:*",
          "logs:*",
          "events:*",
          "dynamodb:*",
          "rds:*",
          "elasticache:*",
          "kms:*"
        ]
        Resource = "*"
      }
    ]
  })
}
