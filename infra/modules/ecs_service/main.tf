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

  # -----------------------------
  # Container Definitions
  # -----------------------------
  container_definitions = jsonencode(
    concat(
      [
        # ---------------------------------------------------------
        # Main Application Container
        # ---------------------------------------------------------
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
            ],
            var.enable_adot ? [
              {
                name  = "OTEL_EXPORTER_OTLP_ENDPOINT"
                value = "http://localhost:4317"
              },
              {
                name  = "OTEL_RESOURCE_ATTRIBUTES"
                value = "service.name=${var.app_name},environment=${var.environment}"
              },
              {
                name  = "OTEL_SERVICE_NAME"
                value = var.app_name
              },
              {
                name  = "OTEL_RESOURCE_ATTRIBUTES"
                value = "service.name=${var.app_name},environment=${var.environment},deployment.environment=${var.environment},source_system=hesp-ecs"
              },
              {
                name  = "OTEL_TRACES_SAMPLER"
                value = "parentbased_traceidratio"
              },
              {
                name  = "OTEL_TRACES_SAMPLER_ARG"
                value = "1.0"
              },
              {
                name  = "OTEL_PROPAGATORS"
                value = "tracecontext,baggage"
              },
              {
                name  = "OTEL_EXPORTER_OTLP_PROTOCOL"
                value = "grpc"
              },
              {
                name  = "OTEL_EXPORTER_OTLP_ENDPOINT"
                value = "http://localhost:4317"
              },
              {
                name  = "DD_API_KEY"
                value = var.dd_api_key
              },
              {
                name  = "HONEYCOMB_API_KEY"
                value = var.honeycomb_api_key
              },
              {
                name  = "HONEYCOMB_DATASET"
                value = var.honeycomb_dataset
              }


            ] : []
          )

          logConfiguration = {
            logDriver = "awslogs"
            options = {
              awslogs-group         = var.log_group_name
              awslogs-region        = "us-east-1"
              awslogs-stream-prefix = var.app_name
            }
          }

          mountPoints = var.enable_adot ? [
            {
              sourceVolume  = "otel-config"
              containerPath = "/etc/otel"
              readOnly      = true
            }
          ] : []
        }
      ],

      # ---------------------------------------------------------
      # ADOT Collector Sidecar (Conditional)
      # ---------------------------------------------------------
      var.enable_adot ? [
        {
          name      = "adot-collector"
          image     = var.adot_image
          essential = true

          portMappings = [
            {
              containerPort = 4317
              protocol      = "tcp"
            },
            {
              containerPort = 4318
              protocol      = "tcp"
            }
          ]

          command = [
            "--config=/etc/otel/collector-config.yaml"
          ]

          mountPoints = [
            {
              sourceVolume  = "otel-config"
              containerPath = "/etc/otel"
              readOnly      = true
            }
          ]

          logConfiguration = {
            logDriver = "awslogs"
            options = {
              awslogs-group         = var.log_group_name
              awslogs-region        = "us-east-1"
              awslogs-stream-prefix = "adot"
            }
          }
        }
      ] : []
    )
  )

  # -----------------------------
  # Volumes (only if ADOT enabled)
  # -----------------------------
  volume {
    name = "otel-config"

    host_path = var.enable_adot ? var.adot_config_file : null
  }

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
    subnets          = var.subnet_ids
    security_groups  = var.security_group_ids
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
