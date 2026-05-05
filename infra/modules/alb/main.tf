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
# ALB
############################################

resource "aws_lb" "this" {
  name               = "${var.app_name}-${var.environment}-alb"
  load_balancer_type = "application"
  security_groups = [var.alb_security_group_id]
  subnets            = var.subnet_ids

  tags = local.base_tags
}

############################################
# Target Group
############################################

resource "aws_lb_target_group" "this" {
  name_prefix        = "hesp-"
  port        = 8080
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = var.vpc_id

  health_check {
    path                = "/healthz"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 2
    matcher             = "200-399"
  }

  tags = local.base_tags
}

############################################
# Listener
############################################

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.this.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.this.arn
  }
}

