############################################
# CloudWatch Log Group (ECS)
############################################

resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/${var.app_name}/${var.environment}/ecs"
  retention_in_days = 30
  tags              = local.base_tags
}

############################################
# ECS Cluster
############################################

resource "aws_ecs_cluster" "cluster" {
  name = "${var.app_name}-${var.environment}-cluster"
  tags = local.base_tags
}

############################################
# ECS Security Group
############################################

resource "aws_security_group" "ecs" {
  name        = "${var.app_name}-${var.environment}-ecs-sg"
  description = "Security group for ECS tasks"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
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
# ECS Service Module
############################################

module "ecs_service" {
  source = "./modules/ecs_service"

  app_name                          = var.app_name
  environment                       = var.environment
  cluster_name                      = aws_ecs_cluster.cluster.name
  compliance_db_host                = var.compliance_db_host
  compliance_db_port                = var.compliance_db_port
  compliance_db_name                = var.compliance_db_name
  compliance_db_username            = var.compliance_db_username
  compliance_db_password_secret_arn = var.compliance_db_password_secret_arn
  container_image                   = var.container_image
  task_execution_role_arn           = module.iam.execution_role_arn
  task_role_arn                     = module.iam.task_role_arn
  subnet_ids                        = module.vpc.private_subnets
  security_group_ids                = [aws_security_group.ecs.id]
  s3_bucket_name                    = aws_s3_bucket.this.bucket
  kms_key_arn                       = aws_kms_key.this.arn
  s3_prefix                         = "events"
  log_group_name                    = aws_cloudwatch_log_group.ecs.name
  desired_count                     = var.desired_count
  target_group_arn                  = module.alb.target_group_arn
  owner                             = var.owner
  cost_center                       = var.cost_center
  tags                              = local.base_tags

  enable_adot = true

  adot_image = "045797643729.dkr.ecr.us-east-1.amazonaws.com/hesp-adot:latest"
}
