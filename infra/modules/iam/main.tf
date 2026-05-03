locals {
  required_tags = {
    environment           = var.environment
    owner                 = var.owner
    "cost-center"         = var.cost_center
    "data-classification" = "phi"
    "managed-by"          = "terraform"
  }

  tags = var.tags

}

# -----------------------------------------------------------------------------
# Execution Role — used by ECS control plane to pull images and write logs
# -----------------------------------------------------------------------------
data "aws_iam_policy_document" "execution_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "execution" {
  name               = "hesp-${var.environment}-ingest-execution"
  assume_role_policy = data.aws_iam_policy_document.execution_assume_role.json
  tags               = local.tags
}

resource "aws_iam_role_policy_attachment" "execution_managed" {
  role       = aws_iam_role.execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# -----------------------------------------------------------------------------
# Task Role — assumed by the running container
# -----------------------------------------------------------------------------
data "aws_iam_policy_document" "task_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "task" {
  name               = "hesp-${var.environment}-ingest-task"
  assume_role_policy = data.aws_iam_policy_document.task_assume_role.json
  tags               = local.tags
}

# -----------------------------------------------------------------------------
# Task Role Policy — least privilege for S3, KMS, Logs, Metrics
# -----------------------------------------------------------------------------
data "aws_iam_policy_document" "task" {
  statement {
    sid      = "AllowS3PutRawEvents"
    effect   = "Allow"
    actions  = ["s3:PutObject"]
    resources = [
      "${var.bucket_arn}/*"
    ]
  }

  statement {
    sid      = "AllowKMSForS3"
    effect   = "Allow"
    actions  = [
      "kms:GenerateDataKey",
      "kms:Decrypt"
    ]
    resources = [var.kms_key_arn]
  }

  statement {
    sid      = "AllowCloudWatchMetrics"
    effect   = "Allow"
    actions  = ["cloudwatch:PutMetricData"]
    resources = ["*"]

    condition {
      test     = "StringEquals"
      variable = "cloudwatch:namespace"
      values   = ["HESP/Ingest"]
    }
  }

  statement {
    sid      = "AllowCloudWatchLogs"
    effect   = "Allow"
    actions  = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["${var.log_group_arn}:*"]
  }
}

resource "aws_iam_policy" "task" {
  name   = "hesp-${var.environment}-ingest-task-policy"
  policy = data.aws_iam_policy_document.task.json
}

resource "aws_iam_role_policy_attachment" "task" {
  role       = aws_iam_role.task.name
  policy_arn = aws_iam_policy.task.arn
}
