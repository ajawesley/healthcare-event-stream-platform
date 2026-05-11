terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

############################################
# DynamoDB Table: compliance_rules
############################################

resource "aws_dynamodb_table" "this" {
  name         = var.table_name
  billing_mode = "PAY_PER_REQUEST"

  hash_key  = "entity_type"
  range_key = "entity_id"

  attribute {
    name = "entity_type"
    type = "S"
  }

  attribute {
    name = "entity_id"
    type = "S"
  }

  ttl {
    enabled        = var.ttl_enabled
    attribute_name = var.ttl_attribute_name
  }

  point_in_time_recovery {
    enabled = var.pitr_enabled
  }

  tags = merge(var.tags, {
    Name = var.table_name
  })
}
