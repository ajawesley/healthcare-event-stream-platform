output "db_endpoint" {
  description = "RDS endpoint"
  value       = aws_db_instance.this.address
}

output "db_port" {
  description = "RDS port"
  value       = aws_db_instance.this.port
}

output "security_group_id" {
  description = "Security group ID for the Compliance DB"
  value       = aws_security_group.this.id
}
