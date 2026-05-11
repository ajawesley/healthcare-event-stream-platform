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
          { name = "S3_BUCKET", value = var.s3_bucket_name },
          { name = "S3_KMS_KEY_ARN", value = var.kms_key_arn },
          { name = "S3_PREFIX", value = var.s3_prefix },

          # -------------------------------
          # Compliance DB Environment Vars
          # -------------------------------
          { name = "COMPLIANCE_DB_HOST", value = var.compliance_db_host },
          { name = "COMPLIANCE_DB_PORT", value = tostring(var.compliance_db_port) },
          { name = "COMPLIANCE_DB_NAME", value = var.compliance_db_name },
          { name = "COMPLIANCE_DB_USER", value = var.compliance_db_username },

          # -------------------------------
          # DynamoDB Compliance Table
          # -------------------------------
          { name = "DYNAMO_TABLE", value = var.dynamodb_table_name },

          # -------------------------------
          # Redis Compliance Cache
          # -------------------------------
          { name = "REDIS_ADDR", value = "${var.redis_primary_endpoint}:6379" }
        ],
        var.enable_adot ? [
          { name = "OTEL_EXPORTER_OTLP_ENDPOINT", value = "localhost:4318" },
          { name = "OTEL_EXPORTER_OTLP_PROTOCOL", value = "http/protobuf" },
          { name = "OTEL_SERVICE_NAME", value = var.app_name },
          { name = "OTEL_PROPAGATORS", value = "tracecontext,baggage" },
          { name = "OTEL_TRACES_SAMPLER", value = "parentbased_traceidratio" },
          { name = "OTEL_TRACES_SAMPLER_ARG", value = "1.0" },
          {
            name  = "OTEL_RESOURCE_ATTRIBUTES",
            value = "service.name=${var.app_name},environment=${var.environment},deployment.environment=${var.environment},source_system=hesp-ecs"
          }
        ] : []
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

    # ---------------------------------------------------------
    # ADOT Collector Sidecar
    # ---------------------------------------------------------
    var.enable_adot ? {
      name      = "adot"
      image     = var.adot_image
      essential = false

      portMappings = [
        { containerPort = 4317, protocol = "tcp" },
        { containerPort = 4318, protocol = "tcp" },
        { containerPort = 13133, protocol = "tcp" }
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
