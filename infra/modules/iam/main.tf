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

############################################
# ECS Execution Role
############################################

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

############################################
# ECS Task Role
############################################

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

############################################
# Glue Job Role (NEW — required for ingestion pipeline)
############################################

data "aws_iam_policy_document" "glue_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["glue.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "glue" {
  name               = "hesp-${var.environment}-glue-role"
  assume_role_policy = data.aws_iam_policy_document.glue_assume_role.json
  tags               = local.tags
}

data "aws_iam_policy_document" "glue_policy" {
  statement {
    sid     = "AllowS3ReadScripts"
    effect  = "Allow"
    actions = ["s3:GetObject"]
    resources = [
      "${var.scripts_bucket_arn}/*"
    ]
  }

  statement {
    sid     = "AllowS3ReadWriteRaw"
    effect  = "Allow"
    actions = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"]
    resources = [
      "${var.bucket_arn}/*"
    ]
  }

  statement {
    sid     = "AllowKMS"
    effect  = "Allow"
    actions = ["kms:Decrypt", "kms:GenerateDataKey"]
    resources = [var.kms_key_arn]
  }

  statement {
    sid     = "AllowCloudWatchLogs"
    effect  = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["${var.log_group_arn}:*"]
  }
}

resource "aws_iam_policy" "glue" {
  name   = "hesp-${var.environment}-glue-policy"
  policy = data.aws_iam_policy_document.glue_policy.json
}

resource "aws_iam_role_policy_attachment" "glue" {
  role       = aws_iam_role.glue.name
  policy_arn = aws_iam_policy.glue.arn
}
