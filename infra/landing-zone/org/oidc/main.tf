############################################
# GitHub OIDC Provider
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
# GitHub Actions Deploy Role (Org-level)
############################################

resource "aws_iam_role" "github_deploy" {
  name = "${var.org_name}-github-oidc-deploy-role"

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
            "token.actions.githubusercontent.com:sub" = [
              "repo:ajawesley/healthcare-event-stream-platform:ref:refs/heads/*",
              "repo:ajawesley/healthcare-event-stream-platform:pull_request",
              "repo:ajawesley/healthcare-event-stream-platform:workflow:*"
            ]
          }
        }
      }
    ]
  })


  tags = merge(
    var.tags,
    {
      "Project" = var.org_name
    }
  )
}

############################################
# Permissions for Terraform Deployments
############################################

resource "aws_iam_role_policy" "github_deploy_policy" {
  name = "${var.org_name}-github-deploy-policy"
  role = aws_iam_role.github_deploy.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          # You can tighten this later; leaving broad for Landing Zone bootstrap
          "ec2:*",
          "s3:*",
          "iam:*",
          "kms:*",
          "lambda:*",
          "cloudwatch:*",
          "logs:*",
          "dynamodb:*",
          "rds:*",
          "elasticloadbalancing:*",
          "ecs:*",
          "events:*",
          "config:*",
          "guardduty:*"
        ]
        Resource = "*"
      }
    ]
  })
}
