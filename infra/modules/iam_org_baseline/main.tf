############################################
# Break-glass Administrator Role
############################################

resource "aws_iam_role" "break_glass" {
  name = "BreakGlassAdmin"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { AWS = var.security_admin_account_id }
      Action    = "sts:AssumeRole"
    }]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "break_glass_admin" {
  role       = aws_iam_role.break_glass.name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
}

############################################
# Audit Read-Only Role
############################################

resource "aws_iam_role" "audit" {
  name = "AuditReadOnly"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { AWS = var.security_admin_account_id }
      Action    = "sts:AssumeRole"
    }]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "audit_security_audit" {
  role       = aws_iam_role.audit.name
  policy_arn = "arn:aws:iam::aws:policy/SecurityAudit"
}

resource "aws_iam_role_policy_attachment" "audit_readonly" {
  role       = aws_iam_role.audit.name
  policy_arn = "arn:aws:iam::aws:policy/ReadOnlyAccess"
}

############################################
# Automation Role (Org-level automation)
############################################

resource "aws_iam_role" "automation" {
  name = "OrgAutomationRole"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { AWS = var.security_admin_account_id }
      Action    = "sts:AssumeRole"
    }]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "automation_admin" {
  role       = aws_iam_role.automation.name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
}

############################################
# Permission Boundary (Optional but recommended)
############################################

resource "aws_iam_policy" "permission_boundary" {
  name        = "OrgPermissionBoundary"
  description = "Permission boundary for all IAM roles created in the organization"
  policy      = file("${path.module}/policies/permission_boundary.json")

  tags = var.tags
}
