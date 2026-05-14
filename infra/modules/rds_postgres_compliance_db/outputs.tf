############################################
# RDS PostgreSQL Outputs
############################################

output "db_endpoint" {
  description = "Primary endpoint address for the Compliance PostgreSQL database"
  value       = aws_db_instance.this.address
}

output "db_port" {
  description = "Port number for the Compliance PostgreSQL database"
  value       = aws_db_instance.this.port
}

output "security_group_id" {
  description = "Security group ID attached to the Compliance DB"
  value       = aws_security_group.this.id
}

output "db_host" {
  description = "Alias for the DB endpoint (same as db_endpoint)"
  value       = aws_db_instance.this.address
}
