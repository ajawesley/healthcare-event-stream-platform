############################################
# Terraform + Provider
############################################

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

############################################
# Common Tags
############################################

locals {
  base_tags = {
    App         = var.app_name
    Environment = var.environment
    Owner       = var.owner
    CostCenter  = var.cost_center
    ManagedBy   = "terraform"
  }
}

############################################
# GitHub OIDC IAM Role
############################################

module "github_oidc" {
  source = "./modules/github-oidc-role"

  role_name    = "github-oidc-deploy-role"
  github_owner = "ajawesley"
  github_repo  = "healthcare-event-stream-platform"
  github_ref   = "*"

  inline_policy_statements = [
    {
      Effect = "Allow"
      Action = [
        "sts:AssumeRole",
        "ecr:*",
        "ecs:*",
        "elasticloadbalancing:*",
        "ec2:*",
        "lambda:*",
        "s3:*",
        "glue:*",
        "cloudwatch:*",
        "logs:*",
        "iam:PassRole"
      ]
      Resource = "*"
    }
  ]
}

############################################
# KMS Key (S3 Encryption)
############################################

resource "aws_kms_key" "this" {
  description             = "KMS key for S3 encryption"
  deletion_window_in_days = 7
  enable_key_rotation     = true
  tags                    = local.base_tags
}

############################################
# Compliance DB Seed Module
############################################
resource "null_resource" "seed_compliance_rules" {
  depends_on = [
    module.compliance_db
  ]

  triggers = {
    db_endpoint = module.compliance_db.db_endpoint
    seed_hash   = filesha256("${path.module}/db/seed_compliance_rules.sql")
  }

  provisioner "local-exec" {
    command = <<EOT
echo "DB_HOST=${var.compliance_db_host}" > seed.env
echo "DB_USER=${var.compliance_db_username}" >> seed.env
echo "DB_PASSWORD=${var.compliance_db_password}" >> seed.env
echo "DB_NAME=${var.compliance_db_name}" >> seed.env
echo "SEED_FILE=${path.module}/db/seed_compliance_rules.sql" >> seed.env

bash ${path.module}/db/run_seed.sh seed.env
EOT
  }



}

