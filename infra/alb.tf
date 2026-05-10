############################################
# ALB Security Group
############################################

resource "aws_security_group" "alb" {
  name        = "${var.app_name}-${var.environment}-alb-sg"
  description = "ALB security group"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.base_tags
}

############################################
# ALB Module
############################################

module "alb" {
  source = "./modules/alb"

  app_name              = var.app_name
  environment           = var.environment
  vpc_id                = module.vpc.vpc_id
  subnet_ids            = module.vpc.public_subnets
  alb_security_group_id = aws_security_group.alb.id

  owner       = var.owner
  cost_center = var.cost_center
  tags        = local.base_tags
}
