locals {
  required_tags = {
    environment           = var.environment
    owner                 = var.owner
    "cost-center"         = var.cost_center
    "data-classification" = "phi"
    "managed-by"          = "terraform"
  }

  tags = merge(var.tags, local.required_tags)
}

# -----------------------------------------------------------------------------
# ECS Cluster
# -----------------------------------------------------------------------------
resource "aws_ecs_cluster" "this" {
  name = var.cluster_name
  tags = local.tags
}

# -----------------------------------------------------------------------------
# ECS Task Definition
# -----------------------------------------------------------------------------
resource "aws_ecs_task_definition" "this" {
  family                   = "${var.app_name}-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = var.task_execution_role_arn
  task_role_arn            = var.task_role_arn

  container_definitions = jsonencode([
    {
      name      = var.app_name
      image     = var.container_image
      essential = true

      portMappings = [
        {
          containerPort = 8080
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "HESP_ENV"
          value = var.environment
        },
        {
          name  = "HESP_S3_BUCKET"
          value = var.s3_bucket_name
        },
        {
          name  = "HESP_LOG_LEVEL"
          value = "info"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = var.log_group_name

          awslogs-region        = data.aws_region.current.name
          awslogs-stream-prefix = var.app_name
        }
      }
    }
  ])

  tags = local.tags
}

data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

# -----------------------------------------------------------------------------
# ECS Service
# -----------------------------------------------------------------------------
resource "aws_ecs_service" "this" {
  name            = "${var.app_name}-svc"
  cluster         = aws_ecs_cluster.this.id
  task_definition = aws_ecs_task_definition.this.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = var.subnet_ids
    security_groups = var.security_group_ids
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = var.target_group_arn
    container_name   = var.app_name
    container_port   = 8080
  }

  lifecycle {
    ignore_changes = [task_definition]
  }

  tags = local.tags
}
