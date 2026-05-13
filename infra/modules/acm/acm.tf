############################################
# Route53 Hosted Zone Lookup
############################################

data "aws_route53_zone" "hesp_root" {
  name         = "hesp.dev."
  private_zone = false
}

############################################
# ACM Certificate for api.hesp.dev
############################################

resource "aws_acm_certificate" "api_hesp" {
  domain_name       = "api.hesp.dev"
  validation_method = "DNS"

  tags = local.base_tags
}

############################################
# DNS Validation Record
############################################

resource "aws_route53_record" "api_hesp_validation" {
  for_each = {
    for dvo in aws_acm_certificate.api_hesp.domain_validation_options :
    dvo.domain_name => {
      name   = dvo.resource_record_name
      type   = dvo.resource_record_type
      record = dvo.resource_record_value
    }
  }

  zone_id = data.aws_route53_zone.hesp_root.zone_id
  name    = each.value.name
  type    = each.value.type
  ttl     = 60
  records = [each.value.record]
}

############################################
# ACM Certificate Validation
############################################

resource "aws_acm_certificate_validation" "api_hesp" {
  certificate_arn         = aws_acm_certificate.api_hesp.arn
  validation_record_fqdns = [for r in aws_route53_record.api_hesp_validation : r.fqdn]
}

############################################
# Output for ALB HTTPS Listener
############################################

output "acm_certificate_arn" {
  value = aws_acm_certificate_validation.api_hesp.certificate_arn
}
