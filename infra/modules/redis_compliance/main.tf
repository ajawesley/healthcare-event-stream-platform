terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

############################################
# Subnet Group (isolated subnets)
############################################

resource "aws_elasticache_subnet_group" "this" {
  name       = "${var.name}-subnet-group"
  subnet_ids = var.isolated_subnet_ids

  description = "Subnet group for compliance Redis"

  tags = merge(var.tags, {
    Name = "${var.name}-subnet-group"
  })
}

############################################
# Security Group
############################################

resource "aws_security_group" "this" {
  name        = "${var.name}-sg"
  description = "Security group for compliance Redis"
  vpc_id      = var.vpc_id

  ingress {
    description = "Allow ECS ingestion service to connect to Redis"
    from_port   = var.port
    to_port     = var.port
    protocol    = "tcp"
    security_groups = [
      var.ingestion_service_sg_id
    ]
  }

  egress {
    description = "Allow all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, {
    Name = "${var.name}-sg"
  })
}

############################################
# Parameter Group
############################################

resource "aws_elasticache_parameter_group" "this" {
  name   = "${var.name}-pg"
  family = var.parameter_group_family

  description = "Parameter group for compliance Redis"

  parameter {
    name  = "maxmemory-policy"
    value = var.maxmemory_policy
  }

  tags = merge(var.tags, {
    Name = "${var.name}-pg"
  })
}

############################################
# Redis Replication Group (Multi-AZ)
############################################

resource "aws_elasticache_replication_group" "this" {
  replication_group_id = var.name
  description          = "Compliance Redis (Multi-AZ)"

  engine         = "redis"
  engine_version = var.engine_version

  node_type            = var.node_type
  port                 = var.port
  parameter_group_name = aws_elasticache_parameter_group.this.name

  subnet_group_name = aws_elasticache_subnet_group.this.name
  security_group_ids = [
    aws_security_group.this.id
  ]

  automatic_failover_enabled = true
  multi_az_enabled           = true

  num_node_groups         = 1
  replicas_per_node_group = var.replicas_per_node_group

  at_rest_encryption_enabled = true
  transit_encryption_enabled = var.transit_encryption_enabled
  auto_minor_version_upgrade = true
  apply_immediately          = true

  tags = merge(var.tags, {
    Name = var.name
  })
}
