locals {
  required_tags = {
    environment           = var.environment
    "data-classification" = "phi"
    owner                 = var.owner
    "cost-center"         = var.cost_center
    "managed-by"          = "terraform"
  }

  tags = merge(var.tags, local.required_tags)
}

# -----------------------------------------------------------------------------
# Security Group
# -----------------------------------------------------------------------------
resource "aws_security_group" "alb" {
  name        = "${var.app_name}-alb-sg"
  description = "ALB security group - HTTP ingest on 80, egress to ECS tasks on 8080 only."
  vpc_id      = var.vpc_id

  lifecycle {
   # create_before_destroy = true
  }

  tags = local.tags
}

resource "aws_vpc_security_group_ingress_rule" "http" {
  security_group_id = aws_security_group.alb.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 80
  to_port           = 80
  ip_protocol       = "tcp"
  description       = "HTTP ingest traffic from all sources."
}

resource "aws_vpc_security_group_egress_rule" "to_ecs" {
  security_group_id            = aws_security_group.alb.id
  referenced_security_group_id = var.ecs_security_group_id
  from_port                    = 8080
  to_port                      = 8080
  ip_protocol                  = "tcp"
  description                  = "Forward traffic to ECS ingest tasks on port 8080 only."
}

# -----------------------------------------------------------------------------
# Application Load Balancer
# -----------------------------------------------------------------------------
resource "aws_lb" "this" {
  name                       = "${var.app_name}-alb"
  load_balancer_type         = "application"
  security_groups            = [aws_security_group.alb.id]
  subnets                    = var.subnet_ids
  internal                   = false
  drop_invalid_header_fields = true
  enable_deletion_protection = var.environment == "prod" ? true : false

  access_logs {
    bucket  = var.access_log_bucket_id
    prefix  = "alb/${var.app_name}"
    enabled = true
  }

  tags = local.tags
}

# -----------------------------------------------------------------------------
# Target Group
# -----------------------------------------------------------------------------
resource "aws_lb_target_group" "this" {
  name        = "${var.app_name}-tg"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "ip"

  health_check {
    path                = "/healthz"
    protocol            = "HTTP"
    matcher             = "200"
    interval            = 15
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
  }

  tags = local.tags
}

# -----------------------------------------------------------------------------
# HTTP Listener — port 80
# -----------------------------------------------------------------------------
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.this.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.this.arn
  }
}
