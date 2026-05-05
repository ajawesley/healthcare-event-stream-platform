locals {
  base_tags = merge(
    var.tags,
    {
      App         = var.app_name
      Environment = var.environment
      Owner       = var.owner
      CostCenter  = var.cost_center
      ManagedBy   = "terraform"
    }
  )
}

resource "aws_ecs_task_definition" "this" {
  family                   = "${var.app_name}-${var.environment}"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "512"
  memory                   = "1024"

  execution_role_arn = var.task_execution_role_arn
  task_role_arn      = var.task_role_arn

  container_definitions = jsonencode([
    {
      name      = var.app_name
      image     = var.container_image
      essential = true

      portMappings = [
        {
          containerPort = 8080
          hostPort      = 8080
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "S3_BUCKET"
          value = var.s3_bucket_name
        },
        {
          name  = "S3_KMS_KEY_ARN"
          value = var.kms_key_arn
        },
        {
          name  = "S3_PREFIX"
          value = var.s3_prefix
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = var.log_group_name
          awslogs-region        = "us-east-1"
          awslogs-stream-prefix = var.app_name
        }
      }
    }
  ])

  tags = local.base_tags
}

resource "aws_ecs_service" "this" {
  name                 = "${var.app_name}-${var.environment}-svc"
  cluster              = var.cluster_name
  task_definition      = aws_ecs_task_definition.this.arn
  force_new_deployment = true
  desired_count        = var.desired_count
  launch_type          = "FARGATE"

  network_configuration {
    subnets         = var.subnet_ids
    security_groups = var.security_group_ids
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = var.target_group_arn
    container_name   = var.app_name
    container_port   = 8080
  }

  lifecycle {
    create_before_destroy = true
  }

  tags = local.base_tags
}
