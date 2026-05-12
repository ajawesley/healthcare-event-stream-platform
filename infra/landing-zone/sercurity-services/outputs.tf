output "guardduty_admin_account_id" {
  description = "Account ID of the GuardDuty delegated administrator"
  value       = var.security_account_id
}

output "securityhub_admin_account_id" {
  description = "Account ID of the Security Hub delegated administrator"
  value       = var.security_account_id
}

output "org_access_analyzer_name" {
  description = "Name of the organization-level IAM Access Analyzer"
  value       = aws_accessanalyzer_analyzer.org.analyzer_name
}

output "guardduty_detector_id" {
  description = "Detector ID for the delegated admin GuardDuty detector"
  value       = aws_guardduty_detector.security_admin.id
}
