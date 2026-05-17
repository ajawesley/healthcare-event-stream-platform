# IaC Guardrails & Paved Roads — Terraform Implementation Guide

This document provides **Terraform guardrail snippets**, **paved‑road examples**, and **module‑level enforcement patterns** for the Healthcare Event Stream Platform (HESP).  
It complements the `iac-architecture.md` by showing **how the guardrails are implemented in code**.

---

# 1. Purpose

This guide ensures:

- consistent Terraform patterns  
- HIPAA‑aligned guardrails  
- paved‑road defaults  
- safe, reproducible infrastructure  
- no “snowflake” resources  

Every module includes:

- **Guardrails** → enforced constraints  
- **Paved Roads** → recommended defaults  
- **Terraform Snippets** → copy‑ready examples  

---

# 2. Core Landing Zone Guardrails

---

## 2.1 VPC Module Guardrails

### Guardrails
- No public IPs  
- Private subnets required  
- Flow logs mandatory  
- NAT gateways only for private egress  
- Default SG locked down  

### Paved Road
All workloads deploy into **private subnets** and use **VPC endpoints**.

### Terraform Snippet — VPC with Guardrails

```hcl
module "vpc" {
  source = "../modules/vpc"

  name = "hesp-dev"

  cidr = "10.10.0.0/16"

  enable_dns_hostnames = true
  enable_dns_support   = true

  public_subnets  = []
  private_subnets = ["10.10.1.0/24", "10.10.2.0/24"]
  isolated_subnets = ["10.10.3.0/24"]

  enable_flow_logs = true

  map_public_ip_on_launch = false

  tags = {
    "phi" = "isolated"
    "env" = "dev"
  }
}
```

---

## 2.2 VPC Endpoints Module Guardrails

### Guardrails
- No internet egress required  
- All AWS API calls routed through endpoints  
- Endpoint SG restricts traffic  

### Paved Road
All ECS, Lambda, Glue workloads **must** use endpoints.

### Terraform Snippet — Required Endpoints

```hcl
module "endpoints" {
  source = "../modules/vpc-endpoints"

  vpc_id             = module.vpc.id
  private_subnet_ids = module.vpc.private_subnet_ids

  endpoints = {
    s3     = { service = "s3" }
    sts    = { service = "sts" }
    ssm    = { service = "ssm" }
    ecr    = { service = "ecr.api" }
    logs   = { service = "logs" }
    secrets = { service = "secretsmanager" }
  }

  endpoint_security_group_id = module.sg.endpoints_sg_id
}
```

---

## 2.3 Security Groups Module Guardrails

### Guardrails
- No `0.0.0.0/0` ingress  
- Egress restricted to AWS endpoints  
- Mandatory descriptions  
- PHI segmentation  

### Paved Road
Use **module‑provided SGs**, not ad‑hoc SGs.

### Terraform Snippet — Least Privilege SG

```hcl
module "sg" {
  source = "../modules/security-groups"

  ecs_ingress = [
    {
      description = "Allow ALB → ECS"
      from_port   = 8080
      to_port     = 8080
      protocol    = "tcp"
      cidr_blocks = []
      security_groups = [module.alb.sg_id]
    }
  ]

  egress = [
    {
      description = "Allow VPC endpoints only"
      from_port   = 443
      to_port     = 443
      protocol    = "tcp"
      security_groups = module.endpoints.sg_ids
    }
  ]
}
```

---

## 2.4 KMS Module Guardrails

### Guardrails
- CMK rotation enabled  
- No wildcard principals  
- Mandatory encryption for all data stores  

### Paved Road
Every module accepts `kms_key_id`.

### Terraform Snippet — CMK with Guardrails

```hcl
resource "aws_kms_key" "main" {
  description         = "HESP CMK"
  enable_key_rotation = true

  policy = data.aws_iam_policy_document.kms.json
}

data "aws_iam_policy_document" "kms" {
  statement {
    sid = "AllowOrgAccounts"
    principals {
      type        = "AWS"
      identifiers = var.allowed_principals
    }
    actions   = ["kms:*"]
    resources = ["*"]
  }
}
```

---

# 3. Governance & Security Guardrails

---

## 3.1 AWS Config Guardrails

### Guardrails
- Required tags  
- Encryption required  
- Public access blocked  
- IAM least privilege  

### Terraform Snippet — Required Tags Rule

```hcl
resource "aws_config_config_rule" "required_tags" {
  name = "required-tags"

  source {
    owner             = "AWS"
    source_identifier = "REQUIRED_TAGS"
  }

  input_parameters = jsonencode({
    tag1Key = "env"
    tag2Key = "phi"
  })
}
```

---

## 3.2 CloudTrail Guardrails

### Guardrails
- Multi‑region  
- Log file validation  
- Centralized log archive  

### Terraform Snippet — Org Trail

```hcl
resource "aws_cloudtrail" "org" {
  name                          = "org-trail"
  is_organization_trail         = true
  is_multi_region_trail         = true
  enable_log_file_validation    = true
  kms_key_id                    = var.kms_key_id
  s3_bucket_name                = var.log_archive_bucket
}
```

---

## 3.3 Security Hub Guardrails

### Guardrails
- CIS AWS Foundations  
- PCI DSS  
- HIPAA Security Rule  

### Terraform Snippet — Auto‑Enable

```hcl
resource "aws_securityhub_account" "main" {}

resource "aws_securityhub_standards_subscription" "cis" {
  standards_arn = "arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
}
```

---

## 3.4 GuardDuty Guardrails

```hcl
resource "aws_guardduty_detector" "main" {
  enable = true
}
```

---

## 3.5 Inspector Guardrails

```hcl
resource "aws_inspector2_enabler" "all" {
  account_ids = [data.aws_caller_identity.current.account_id]
  resource_types = ["EC2", "ECR", "LAMBDA"]
}
```

---

## 3.6 SCP Baseline Guardrails

### Guardrails
- Deny public S3  
- Deny IAM wildcard  
- Deny unencrypted resources  
- Deny disabling CloudTrail/Config  

### Terraform Snippet — Deny Public S3

```hcl
statement {
  sid = "DenyPublicS3"
  effect = "Deny"

  actions   = ["s3:PutBucketAcl", "s3:PutBucketPolicy"]
  resources = ["arn:aws:s3:::*"]

  condition {
    test     = "StringEquals"
    variable = "s3:x-amz-acl"
    values   = ["public-read", "public-read-write"]
  }
}
```

---

# 4. Data Plane Guardrails

---

## 4.1 S3 Buckets Module Guardrails

### Guardrails
- Versioning  
- Encryption  
- Block public access  
- Lifecycle policies  
- Access logging  

### Terraform Snippet — Raw Bucket

```hcl
module "s3_raw" {
  source = "../modules/s3"

  bucket_name = "hesp-raw-${var.env}"

  versioning = true
  block_public_access = true
  enable_access_logging = true

  lifecycle_rules = [
    {
      id      = "expire-old"
      enabled = true
      expiration = { days = 365 }
    }
  ]

  kms_key_id = var.kms_key_id
}
```

---

## 4.2 RDS Guardrails

```hcl
module "rds" {
  source = "../modules/rds"

  engine               = "postgres"
  multi_az             = true
  storage_encrypted    = true
  kms_key_id           = var.kms_key_id
  deletion_protection  = true
  performance_insights = true
}
```

---

## 4.3 DynamoDB Guardrails

```hcl
resource "aws_dynamodb_table" "rules" {
  name         = "compliance-rules"
  billing_mode = "PAY_PER_REQUEST"

  point_in_time_recovery {
    enabled = true
  }

  server_side_encryption {
    enabled     = true
    kms_key_arn = var.kms_key_id
  }
}
```

---

## 4.4 Redis Guardrails

```hcl
module "redis" {
  source = "../modules/redis"

  transit_encryption_enabled = true
  at_rest_encryption_enabled = true
  auth_token                 = var.redis_auth_token
}
```

---

# 5. Application Runtime Guardrails

---

## 5.1 ECS Cluster Guardrails

```hcl
module "ecs_cluster" {
  source = "../modules/ecs-cluster"

  fargate_only = true
  container_insights = true
}
```

---

## 5.2 ECS Service Guardrails

```hcl
module "ecs_service" {
  source = "../modules/ecs-service"

  cpu    = 512
  memory = 1024

  readonly_root_filesystem = true
  assign_public_ip         = false

  health_check = {
    command = ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"]
  }
}
```

---

## 5.3 Lambda Trigger Guardrails

```hcl
module "lambda_trigger" {
  source = "../modules/lambda-trigger"

  vpc_subnet_ids = module.vpc.private_subnet_ids
  security_group_ids = [module.sg.lambda_sg_id]

  dead_letter_queue_arn = module.sqs.dlq_arn
  kms_key_id            = var.kms_key_id

  environment = {
    variables = {
      STAGE = var.env
    }
  }
}
```

---

## 5.4 Glue Guardrails

```hcl
module "glue_job" {
  source = "../modules/glue-job"

  security_configuration = module.glue_security_config.name
  kms_key_id             = var.kms_key_id
}
```

---

# 6. CI/CD Guardrails

---

## 6.1 Required Pipeline Stages

```yaml
stages:
  - validate
  - plan
  - policy-check
  - apply
  - verify
```

---

## 6.2 Policy Enforcement (OPA Example)

```rego
deny[msg] {
  input.resource.type == "aws_s3_bucket"
  input.resource.public == true
  msg = "Public S3 buckets are not allowed"
}
```

---

# 7. Outcomes

- Guardrails enforced in Terraform  
- Paved roads for all modules  
- HIPAA‑aligned defaults  
- No snowflake infrastructure  
- Safe, observable deployments  
- Consistent environments across dev/stage/prod  
