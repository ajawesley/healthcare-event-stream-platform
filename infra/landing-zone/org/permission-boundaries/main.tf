terraform {
  required_version = ">= 1.5.0"
}

locals {
  policy_path = "${path.module}/policies/hesp-permission-boundary.json"
}

resource "aws_iam_policy" "hesp_permission_boundary" {
  name        = "hesp-permission-boundary"
  description = "Least privilege envelope for workloads"
  policy      = file(local.policy_path)
}
