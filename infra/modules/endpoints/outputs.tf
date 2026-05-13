output "vpce_security_group_id" {
  description = "Security group ID for interface endpoints"
  value       = aws_security_group.vpce.id
}

output "gateway_endpoints" {
  description = "Gateway endpoint IDs"
  value = {
    s3       = aws_vpc_endpoint.s3.id
    dynamodb = aws_vpc_endpoint.dynamodb.id
  }
}

output "interface_endpoints" {
  description = "Interface endpoint IDs"
  value       = aws_vpc_endpoint.interface
}
