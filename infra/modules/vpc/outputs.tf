output "vpc_id" {
  value = aws_vpc.this.id
}

output "vpc_cidr" {
  value = aws_vpc.this.cidr_block
}

output "public_subnets" {
  value = values(aws_subnet.public)[*].id
}

output "private_subnets" {
  value = values(aws_subnet.private)[*].id
}

output "nat_gateway_id" {
  value = aws_nat_gateway.nat.id
}

output "endpoint_sg_id" {
  value = aws_security_group.endpoints.id
}
