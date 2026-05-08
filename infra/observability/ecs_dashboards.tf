resource "aws_cloudwatch_dashboard" "ecs_service_observability" {
  dashboard_name = "hesp-${var.environment}-ecs-service-observability"

  dashboard_body = jsonencode({
    widgets = [
      # ---------------------------------------------------------------
      # Header
      # ---------------------------------------------------------------
      {
        "type" : "text",
        "x" : 0,
        "y" : 0,
        "width" : 24,
        "height" : 2,
        "properties" : {
          "markdown" : "# HESP ${var.environment} – ECS Service & API Observability\nCPU, memory, restarts, ALB 4xx/5xx, latency, error rate, and saturation."
        }
      },

      # ---------------------------------------------------------------
      # ECS Service Health – CPU / Memory
      # ---------------------------------------------------------------
      {
        "type" : "metric",
        "x" : 0,
        "y" : 2,
        "width" : 12,
        "height" : 6,
        "properties" : {
          "title" : "ECS Service CPU Utilization",
          "view" : "timeSeries",
          "stacked" : false,
          "metrics" : [
            [
              "ECS/ContainerInsights",
              "CpuUtilized",
              "ClusterName", aws_ecs_cluster.cluster.name,
              "ServiceName", "${var.app_name}-${var.environment}-svc"
            ],
            [
              ".",
              "CpuReserved",
              "ClusterName", aws_ecs_cluster.cluster.name,
              "ServiceName", "${var.app_name}-${var.environment}-svc"
            ]
          ],
          "region" : var.aws_region,
          "stat" : "Average",
          "period" : 60
        }
      },
      {
        "type" : "metric",
        "x" : 12,
        "y" : 2,
        "width" : 12,
        "height" : 6,
        "properties" : {
          "title" : "ECS Service Memory Utilization",
          "view" : "timeSeries",
          "stacked" : false,
          "metrics" : [
            [
              "ECS/ContainerInsights",
              "MemoryUtilized",
              "ClusterName", aws_ecs_cluster.cluster.name,
              "ServiceName", "${var.app_name}-${var.environment}-svc"
            ],
            [
              ".",
              "MemoryReserved",
              "ClusterName", aws_ecs_cluster.cluster.name,
              "ServiceName", "${var.app_name}-${var.environment}-svc"
            ]
          ],
          "region" : var.aws_region,
          "stat" : "Average",
          "period" : 60
        }
      },

      # ---------------------------------------------------------------
      # ECS Tasks / Restarts / Deployment Stability
      # ---------------------------------------------------------------
      {
        "type" : "metric",
        "x" : 0,
        "y" : 8,
        "width" : 12,
        "height" : 6,
        "properties" : {
          "title" : "ECS Tasks (Desired vs Running)",
          "view" : "timeSeries",
          "stacked" : false,
          "metrics" : [
            [
              "ECS/ContainerInsights",
              "DesiredTaskCount",
              "ClusterName", aws_ecs_cluster.cluster.name,
              "ServiceName", "${var.app_name}-${var.environment}-svc"
            ],
            [
              ".",
              "RunningTaskCount",
              "ClusterName", aws_ecs_cluster.cluster.name,
              "ServiceName", "${var.app_name}-${var.environment}-svc"
            ]
          ],
          "region" : var.aws_region,
          "stat" : "Average",
          "period" : 60
        }
      },
      {
        "type" : "metric",
        "x" : 12,
        "y" : 8,
        "width" : 12,
        "height" : 6,
        "properties" : {
          "title" : "Task Restarts / OOM Kills",
          "view" : "timeSeries",
          "stacked" : false,
          "metrics" : [
            [
              "ECS/ContainerInsights",
              "TaskRestartCount",
              "ClusterName", aws_ecs_cluster.cluster.name,
              "ServiceName", "${var.app_name}-${var.environment}-svc"
            ],
            [
              ".",
              "ContainerExitCodeCount",
              "ClusterName", aws_ecs_cluster.cluster.name,
              "ServiceName", "${var.app_name}-${var.environment}-svc",
              "ExitCode", "137"
            ]
          ],
          "region" : var.aws_region,
          "stat" : "Sum",
          "period" : 300
        }
      },

      # ---------------------------------------------------------------
      # ALB – 4xx / 5xx / Latency
      # (assumes module.alb exposes alb_arn_suffix and target_group_arn_suffix)
      # ---------------------------------------------------------------
      {
        "type" : "metric",
        "x" : 0,
        "y" : 14,
        "width" : 12,
        "height" : 6,
        "properties" : {
          "title" : "ALB 4xx / 5xx Error Rates",
          "view" : "timeSeries",
          "stacked" : false,
          "metrics" : [
            [
              "AWS/ApplicationELB",
              "HTTPCode_Target_4XX_Count",
              "LoadBalancer", module.alb.alb_arn_suffix
            ],
            [
              ".",
              "HTTPCode_Target_5XX_Count",
              "LoadBalancer", module.alb.alb_arn_suffix
            ]
          ],
          "region" : var.aws_region,
          "stat" : "Sum",
          "period" : 60
        }
      },
      {
        "type" : "metric",
        "x" : 12,
        "y" : 14,
        "width" : 12,
        "height" : 6,
        "properties" : {
          "title" : "ALB Target Response Time (p50/p95/p99)",
          "view" : "timeSeries",
          "stacked" : false,
          "metrics" : [
            [
              "AWS/ApplicationELB",
              "TargetResponseTime",
              "TargetGroup", module.alb.target_group_arn_suffix,
              "LoadBalancer", module.alb.alb_arn_suffix,
              { "stat" : "p50" }
            ],
            [
              ".",
              "TargetResponseTime",
              "TargetGroup", module.alb.target_group_arn_suffix,
              "LoadBalancer", module.alb.alb_arn_suffix,
              { "stat" : "p95" }
            ],
            [
              ".",
              "TargetResponseTime",
              "TargetGroup", module.alb.target_group_arn_suffix,
              "LoadBalancer", module.alb.alb_arn_suffix,
              { "stat" : "p99" }
            ]
          ],
          "region" : var.aws_region,
          "stat" : "Average",
          "period" : 60
        }
      },

      # ---------------------------------------------------------------
      # API Performance – Latency / Error Rate / Throughput (OTEL)
      # ---------------------------------------------------------------
      {
        "type" : "metric",
        "x" : 0,
        "y" : 20,
        "width" : 12,
        "height" : 6,
        "properties" : {
          "title" : "API Latency (p50/p95/p99)",
          "view" : "timeSeries",
          "stacked" : false,
          "metrics" : [
            [
              "HESP/API",
              "http_server_duration_p50",
              "Environment", var.environment,
              "Service", "${var.app_name}-${var.environment}-svc"
            ],
            [
              ".",
              "http_server_duration_p95",
              "Environment", var.environment,
              "Service", "${var.app_name}-${var.environment}-svc"
            ],
            [
              ".",
              "http_server_duration_p99",
              "Environment", var.environment,
              "Service", "${var.app_name}-${var.environment}-svc"
            ]
          ],
          "region" : var.aws_region,
          "stat" : "Average",
          "period" : 60
        }
      },
      {
        "type" : "metric",
        "x" : 12,
        "y" : 20,
        "width" : 12,
        "height" : 6,
        "properties" : {
          "title" : "API Error Rate (4xx/5xx) & Throughput",
          "view" : "timeSeries",
          "stacked" : false,
          "metrics" : [
            [
              "HESP/API",
              "http_server_requests_total",
              "Environment", var.environment,
              "Service", "${var.app_name}-${var.environment}-svc",
              "status_code_class", "2xx"
            ],
            [
              ".",
              "http_server_requests_total",
              "Environment", var.environment,
              "Service", "${var.app_name}-${var.environment}-svc",
              "status_code_class", "4xx"
            ],
            [
              ".",
              "http_server_requests_total",
              "Environment", var.environment,
              "Service", "${var.app_name}-${var.environment}-svc",
              "status_code_class", "5xx"
            ]
          ],
          "region" : var.aws_region,
          "stat" : "Sum",
          "period" : 60
        }
      },

      # ---------------------------------------------------------------
      # Saturation – DB / Redis / Queue Depth (OTEL custom metrics)
      # ---------------------------------------------------------------
      {
        "type" : "metric",
        "x" : 0,
        "y" : 26,
        "width" : 12,
        "height" : 6,
        "properties" : {
          "title" : "DB Connection Pool Usage",
          "view" : "timeSeries",
          "stacked" : false,
          "metrics" : [
            [
              "HESP/API",
              "db_client_connections_usage",
              "Environment", var.environment,
              "Service", "${var.app_name}-${var.environment}-svc"
            ]
          ],
          "region" : var.aws_region,
          "stat" : "Average",
          "period" : 60
        }
      },
      {
        "type" : "metric",
        "x" : 12,
        "y" : 26,
        "width" : 12,
        "height" : 6,
        "properties" : {
          "title" : "Redis / Queue Saturation",
          "view" : "timeSeries",
          "stacked" : false,
          "metrics" : [
            [
              "HESP/API",
              "redis_client_connections_usage",
              "Environment", var.environment,
              "Service", "${var.app_name}-${var.environment}-svc"
            ],
            [
              ".",
              "queue_depth",
              "Environment", var.environment,
              "Service", "${var.app_name}-${var.environment}-svc"
            ]
          ],
          "region" : var.aws_region,
          "stat" : "Average",
          "period" : 60
        }
      }
    ]
  })
}
