locals {
  scp_files = fileset(var.scp_directory, "*.json")
}

resource "aws_organizations_policy" "scp" {
  for_each = { for file in local.scp_files : file => file }

  name        = replace(basename(each.key), ".json", "")
  description = "SCP: ${replace(basename(each.key), ".json", "")}"
  type        = "SERVICE_CONTROL_POLICY"
  content     = file("${var.scp_directory}/${each.key}")
}

resource "aws_organizations_policy_attachment" "ou_attachment" {
  for_each = aws_organizations_policy.scp

  policy_id = each.value.id
  target_id = var.ou_id
}
