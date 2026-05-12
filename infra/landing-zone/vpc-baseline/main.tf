terraform {
  required_version = ">= 1.5.0"
}

provider "aws" {
  region = var.aws_region
}

locals {
  vpc_cidr = var.vpc_cidr
  azs      = var.azs

  tags = merge(
    {
      Project = var.org_name
      Owner   = var.owner
    },
    var.extra_tags
  )
}

############################################
# VPC
############################################

resource "aws_vpc" "baseline" {
  cidr_block           = local.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = merge(local.tags, {
    Name = "${var.org_name}-baseline-vpc"
  })
}

############################################
# Internet Gateway
############################################

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.baseline.id

  tags = merge(local.tags, {
    Name = "${var.org_name}-baseline-igw"
  })
}

############################################
# Subnets
############################################

resource "aws_subnet" "public" {
  for_each = { for idx, az in local.azs : az => idx }

  vpc_id                  = aws_vpc.baseline.id
  cidr_block              = cidrsubnet(local.vpc_cidr, 4, each.value)
  availability_zone       = each.key
  map_public_ip_on_launch = true

  tags = merge(local.tags, {
    Name = "${var.org_name}-baseline-public-${each.key}"
  })
}

resource "aws_subnet" "private" {
  for_each = { for idx, az in local.azs : az => idx }

  vpc_id            = aws_vpc.baseline.id
  cidr_block        = cidrsubnet(local.vpc_cidr, 4, each.value + 8)
  availability_zone = each.key

  tags = merge(local.tags, {
    Name = "${var.org_name}-baseline-private-${each.key}"
  })
}

############################################
# NAT Gateway (single AZ for cost optimization)
############################################

resource "aws_eip" "nat" {
  domain = "vpc"

  tags = merge(local.tags, {
    Name = "${var.org_name}-baseline-nat-eip"
  })
}

resource "aws_nat_gateway" "nat" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public[local.azs[0]].id

  tags = merge(local.tags, {
    Name = "${var.org_name}-baseline-nat"
  })
}

############################################
# Route Tables
############################################

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.baseline.id

  tags = merge(local.tags, {
    Name = "${var.org_name}-baseline-public-rt"
  })
}

resource "aws_route" "public_internet" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.igw.id
}

resource "aws_route_table_association" "public" {
  for_each = aws_subnet.public

  subnet_id      = each.value.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.baseline.id

  tags = merge(local.tags, {
    Name = "${var.org_name}-baseline-private-rt"
  })
}

resource "aws_route" "private_nat" {
  route_table_id         = aws_route_table.private.id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.nat.id
}

resource "aws_route_table_association" "private" {
  for_each = aws_subnet.private

  subnet_id      = each.value.id
  route_table_id = aws_route_table.private.id
}

############################################
# VPC Endpoints
############################################

# Gateway endpoints
resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.baseline.id
  service_name = "com.amazonaws.${var.aws_region}.s3"
  vpc_endpoint_type = "Gateway"

  route_table_ids = [
    aws_route_table.private.id
  ]

  tags = merge(local.tags, {
    Name = "${var.org_name}-vpce-s3"
  })
}

resource "aws_vpc_endpoint" "dynamodb" {
  vpc_id       = aws_vpc.baseline.id
  service_name = "com.amazonaws.${var.aws_region}.dynamodb"
  vpc_endpoint_type = "Gateway"

  route_table_ids = [
    aws_route_table.private.id
  ]

  tags = merge(local.tags, {
    Name = "${var.org_name}-vpce-dynamodb"
  })
}

# Interface endpoints
locals {
  interface_services = [
    "logs",
    "sts",
    "ecr.api",
    "ecr.dkr",
    "ssm",
    "ssmmessages",
    "ec2messages",
    "secretsmanager",
    "kms"
  ]
}

resource "aws_vpc_endpoint" "interface" {
  for_each = toset(local.interface_services)

  vpc_id              = aws_vpc.baseline.id
  service_name        = "com.amazonaws.${var.aws_region}.${each.key}"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = [for s in aws_subnet.private : s.id]
  private_dns_enabled = true

  security_group_ids = []

  tags = merge(local.tags, {
    Name = "${var.org_name}-vpce-${each.key}"
  })
}

############################################
# VPC Flow Logs → Central Log Archive
############################################

resource "aws_flow_log" "vpc" {
  log_destination      = var.log_archive_bucket_arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.baseline.id

  tags = merge(local.tags, {
    Name = "${var.org_name}-baseline-flowlogs"
  })
}
