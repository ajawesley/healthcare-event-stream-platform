############################################
# GitHub OIDC Provider (Org Account)
############################################

resource "aws_iam_openid_connect_provider" "github" {
  url = "https://token.actions.githubusercontent.com"

  client_id_list = [
    "sts.amazonaws.com"
  ]

  thumbprint_list = [
    "6938fd4d98bab03faadb97b34396831e3780aea1"
  ]
}

############################################
# GitHub OIDC Role for Landing Zone Phase 1 (Org)
############################################

resource "aws_iam_role" "github_oidc_org" {
  name = "github-oidc-org"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = aws_iam_openid_connect_provider.github.arn
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

resource "aws_iam_role_policy" "github_oidc_org_policy" {
  role = aws_iam_role.github_oidc_org.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "organizations:*",
          "iam:*",
          "cloudtrail:*",
          "config:*",
          "guardduty:*",
          "securityhub:*",
          "s3:*",
          "kms:*"
        ]
        Resource = "*"
      }
    ]
  })
}
