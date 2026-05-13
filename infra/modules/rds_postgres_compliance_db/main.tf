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
# Subnet Group (isolated only)
############################################

resource "aws_db_subnet_group" "this" {
  name       = "${var.name}-subnet-group"
  subnet_ids = var.isolated_subnet_ids

  tags = merge(var.tags, {
    Name = "${var.name}-subnet-group"
  })
}

############################################
# Security Group
############################################

resource "aws_security_group" "this" {
  name        = "${var.name}-sg"
  description = "Security group for Compliance DB"
  vpc_id      = var.vpc_id

  ingress {
    description = "Allow ingestion service to connect"
    from_port   = 5432
    to_port     = 5432
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

resource "aws_db_parameter_group" "this" {
  name        = "${var.name}-pg"
  family      = "postgres14"
  description = "Parameter group for Compliance DB"

  tags = merge(var.tags, {
    Name = "${var.name}-pg"
  })
}

############################################
# RDS PostgreSQL Instance
############################################

resource "aws_db_instance" "this" {
  identifier = var.name

  engine         = "postgres"
  engine_version = var.engine_version

  instance_class        = var.instance_class
  allocated_storage     = var.allocated_storage
  max_allocated_storage = var.max_allocated_storage

  db_name  = var.db_name
  username = var.db_username
  password = var.db_password

  db_subnet_group_name   = aws_db_subnet_group.this.name
  vpc_security_group_ids = [aws_security_group.this.id]

  publicly_accessible = false
  multi_az            = var.multi_az
  storage_encrypted   = true

  backup_retention_period = var.backup_retention_days
  skip_final_snapshot     = true
  deletion_protection     = false

  parameter_group_name = aws_db_parameter_group.this.name

  tags = merge(var.tags, {
    Name = "${var.name}"
  })
}
