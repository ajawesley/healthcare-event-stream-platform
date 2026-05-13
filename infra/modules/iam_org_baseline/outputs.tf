output "break_glass_role_arn" {
  value = aws_iam_role.break_glass.arn
}

output "audit_role_arn" {
  value = aws_iam_role.audit.arn
}

output "automation_role_arn" {
  value = aws_iam_role.automation.arn
}

output "permission_boundary_arn" {
  value = aws_iam_policy.permission_boundary.arn
}
