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

############################################
# ECS Task Definition
############################################

resource "aws_ecs_task_definition" "this" {
  family                   = "${var.app_name}-${var.environment}"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "512"
  memory                   = "1024"

  execution_role_arn = var.task_execution_role_arn
  task_role_arn      = var.task_role_arn

  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "X86_64"
  }

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

      environment = concat(
        [
          { name = "REVISION_FORCE", value = "45" },

          { name = "S3_BUCKET", value = var.s3_bucket_name },
          { name = "S3_KMS_KEY_ARN", value = var.kms_key_arn },
          { name = "S3_PREFIX", value = var.s3_prefix },

          { name = "COMPLIANCE_DB_HOST", value = var.compliance_db_host },
          { name = "COMPLIANCE_DB_PORT", value = tostring(var.compliance_db_port) },
          { name = "COMPLIANCE_DB_NAME", value = var.compliance_db_name },
          { name = "COMPLIANCE_DB_USER", value = var.compliance_db_username },

          { name = "DYNAMO_TABLE", value = var.dynamodb_table_name },

          { name = "REDIS_ADDR", value = "${var.redis_primary_endpoint}:6379" },

          # OTEL → ADOT sidecar
          { name = "OTEL_EXPORTER_OTLP_ENDPOINT", value = "http://127.0.0.1:4318" },
          { name = "OTEL_SERVICE_NAME", value = var.app_name },
          { name = "OTEL_RESOURCE_ATTRIBUTES", value = "service.name=${var.app_name},service.version=v1.0.0,environment=${var.environment}" }
        ],
        []
      )

      secrets = [
        {
          name      = "COMPLIANCE_DB_PASSWORD"
          valueFrom = var.compliance_db_password_secret_arn
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
    },

    ###############################################################
    # ADOT Collector Sidecar (ENABLED)
    ###############################################################
    var.enable_adot ? {
      name      = "adot"
      image     = var.adot_image
      essential = false

      portMappings = [
        { containerPort = 4317, protocol = "tcp" }, # gRPC
        { containerPort = 4318, protocol = "tcp" }, # HTTP
        { containerPort = 13133, protocol = "tcp" } # health check
      ]

      environment = [
        { name = "AWS_REGION", value = "us-east-1" }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = var.log_group_name
          awslogs-region        = "us-east-1"
          awslogs-stream-prefix = "adot"
        }
      }
    } : null
  ])

  tags = local.base_tags
}

############################################
# ECS Service
############################################

resource "aws_ecs_service" "this" {
  name                 = "${var.app_name}-${var.environment}-svc"
  cluster              = var.cluster_name
  task_definition      = aws_ecs_task_definition.this.arn
  force_new_deployment = true
  desired_count        = var.desired_count
  launch_type          = "FARGATE"

  enable_execute_command = true

  network_configuration {
    subnets          = var.subnet_ids
    security_groups  = var.security_group_ids
    assign_public_ip = false
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
