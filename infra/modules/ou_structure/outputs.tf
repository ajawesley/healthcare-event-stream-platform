output "ou_ids" {
  description = "Map of OU names to IDs"
  value = {
    security       = aws_organizations_organizational_unit.security.id
    infrastructure = aws_organizations_organizational_unit.infrastructure.id
    workloads      = aws_organizations_organizational_unit.workloads.id
    sandbox        = aws_organizations_organizational_unit.sandbox.id
    dev            = aws_organizations_organizational_unit.dev.id
    qa             = aws_organizations_organizational_unit.qa.id
    prod           = aws_organizations_organizational_unit.prod.id
  }
}
