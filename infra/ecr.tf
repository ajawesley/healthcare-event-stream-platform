############################################
# ECR Repository for Ingest Service
############################################

resource "aws_ecr_repository" "ingest" {
  name = "${var.app_name}-${var.environment}-ingest"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = local.base_tags
}

############################################
# CodeBuild IAM Role
############################################

resource "aws_iam_role" "codebuild_ingest_role" {
  name = "${var.app_name}-${var.environment}-codebuild-ingest-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "codebuild.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })

  tags = local.base_tags
}

resource "aws_iam_role_policy" "codebuild_ingest_policy" {
  role = aws_iam_role.codebuild_ingest_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:CompleteLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:InitiateLayerUpload",
          "ecr:PutImage"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "*"
      }
    ]
  })
}

############################################
# CodeBuild Project for Ingest Service
############################################

resource "aws_codebuild_project" "ingest" {
  name          = "${var.app_name}-${var.environment}-ingest-build"
  description   = "Builds and pushes ingest-service Docker image to ECR"
  service_role  = aws_iam_role.codebuild_ingest_role.arn
  build_timeout = 30

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type    = "BUILD_GENERAL1_SMALL"
    image           = "aws/codebuild/standard:7.0"
    type            = "LINUX_CONTAINER"
    privileged_mode = true

    environment_variable {
      name  = "ECR_REPO"
      value = aws_ecr_repository.ingest.repository_url
    }
  }

  source {
    type            = "GITHUB"
    location        = "https://github.com/ajawesley/healthcare-event-stream-platform"
    git_clone_depth = 1
    buildspec       = "buildspec-ingest.yml"
  }

  tags = local.base_tags
}

############################################
# Outputs
############################################

output "ingest_ecr_repo_url" {
  value = aws_ecr_repository.ingest.repository_url
}

output "ingest_codebuild_project_name" {
  value = aws_codebuild_project.ingest.name
}