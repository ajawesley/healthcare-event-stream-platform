locals {
  alb_arn_suffix          = regex("app/(.*)", var.alb_arn)[0]
  target_group_arn_suffix = regex("targetgroup/(.*)", var.target_group_arn)[0]
}

resource "aws_cloudwatch_dashboard" "ecs_service_observability" {
  dashboard_name = "hesp-${var.environment}-ecs-service-observability"

  dashboard_body = jsonencode({
    widgets = [
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

      # ECS CPU
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
              "ClusterName", var.ecs_cluster_name,
              "ServiceName", var.ecs_service_name
            ],
            [
              ".",
              "CpuReserved",
              "ClusterName", var.ecs_cluster_name,
              "ServiceName", var.ecs_service_name
            ]
          ],
          "region" : var.aws_region,
          "stat" : "Average",
          "period" : 60
        }
      },

      # ECS Memory
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
              "ClusterName", var.ecs_cluster_name,
              "ServiceName", var.ecs_service_name
            ],
            [
              ".",
              "MemoryReserved",
              "ClusterName", var.ecs_cluster_name,
              "ServiceName", var.ecs_service_name
            ]
          ],
          "region" : var.aws_region,
          "stat" : "Average",
          "period" : 60
        }
      },

      # ECS Tasks
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
              "ClusterName", var.ecs_cluster_name,
              "ServiceName", var.ecs_service_name
            ],
            [
              ".",
              "RunningTaskCount",
              "ClusterName", var.ecs_cluster_name,
              "ServiceName", var.ecs_service_name
            ]
          ],
          "region" : var.aws_region,
          "stat" : "Average",
          "period" : 60
        }
      },

      # ECS Restarts / OOM
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
              "ClusterName", var.ecs_cluster_name,
              "ServiceName", var.ecs_service_name
            ],
            [
              ".",
              "ContainerExitCodeCount",
              "ClusterName", var.ecs_cluster_name,
              "ServiceName", var.ecs_service_name,
              "ExitCode", "137"
            ]
          ],
          "region" : var.aws_region,
          "stat" : "Sum",
          "period" : 300
        }
      },

      # ALB 4xx/5xx
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
              "LoadBalancer", local.alb_arn_suffix
            ],
            [
              ".",
              "HTTPCode_Target_5XX_Count",
              "LoadBalancer", local.alb_arn_suffix
            ]
          ],
          "region" : var.aws_region,
          "stat" : "Sum",
          "period" : 60
        }
      },

      # ALB Latency
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
              "TargetGroup", local.target_group_arn_suffix,
              "LoadBalancer", local.alb_arn_suffix,
              { "stat" : "p50" }
            ],
            [
              ".",
              "TargetResponseTime",
              "TargetGroup", local.target_group_arn_suffix,
              "LoadBalancer", local.alb_arn_suffix,
              { "stat" : "p95" }
            ],
            [
              ".",
              "TargetResponseTime",
              "TargetGroup", local.target_group_arn_suffix,
              "LoadBalancer", local.alb_arn_suffix,
              { "stat" : "p99" }
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
