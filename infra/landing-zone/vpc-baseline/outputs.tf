output "vpc_id" {
  description = "ID of the baseline VPC"
  value       = aws_vpc.baseline.id
}

output "public_subnet_ids" {
  description = "List of public subnet IDs"
  value       = [for s in aws_subnet.public : s.id]
}

output "private_subnet_ids" {
  description = "List of private subnet IDs"
  value       = [for s in aws_subnet.private : s.id]
}

output "nat_gateway_id" {
  description = "ID of the NAT Gateway"
  value       = aws_nat_gateway.nat.id
}

output "vpc_endpoint_ids" {
  description = "Map of VPC endpoint IDs"
  value       = {
    s3       = aws_vpc_endpoint.s3.id
    dynamodb = aws_vpc_endpoint.dynamodb.id
    interface = { for k, v in aws_vpc_endpoint.interface : k => v.id }
  }
}
