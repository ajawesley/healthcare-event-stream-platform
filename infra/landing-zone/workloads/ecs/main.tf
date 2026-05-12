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
# ECS Task Assume Role Policy
# ---------------------------------------------------------
data "aws_iam_policy_document" "ecs_task_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

# ---------------------------------------------------------
# ECS Task Role
# ---------------------------------------------------------
resource "aws_iam_role" "ecs_task_role" {
  name               = "hesp-ecs-task-role"
  assume_role_policy = data.aws_iam_policy_document.ecs_task_assume.json

  # Enforced by SCP — must be present
  permissions_boundary = module.iam_boundaries.permission_boundary_arn

  tags = var.tags
}

# ---------------------------------------------------------
# Attach AWS-managed ECS policies
# ---------------------------------------------------------
resource "aws_iam_role_policy_attachment" "ecs_task_execution" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}
