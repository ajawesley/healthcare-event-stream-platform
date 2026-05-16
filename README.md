```md
# AcmeCo Event Stream Platform (ESP)

A secure, compliant, healthcare‑grade event streaming and data platform designed to:

- Ingest and validate events  
- Canonicalize heterogeneous payloads  
- Enforce compliance (HIPAA + internal policies)  
- Store data in a governed, KMS‑encrypted data lake  
- Support analytics, ML, and replay workflows  
- Provide clean integration surfaces for producers and consumers  

---

## Repository Structure

```
infra/
  org-management/        # Org-level: OUs, SCPs, CloudTrail, Config, log archive
  envs/
    dev/
      main.tf            # Dev environment: VPC, ECS, RDS, Redis, S3, Glue, Config
  modules/
    vpc/
    endpoints/
    s3_buckets/
    s3_log_archive/
    glue_job/
    glue_crawlers/
    lambda_trigger/
    dynamodb_compliance_rules/
    redis_compliance/
    rds_postgres_compliance_db/
    ecs_service/
    alb/
    iam/
    config/
    kms_cloudtrail/
    cloudtrail_org/
    centralized_logging/
    ou_structure/
    scp_baseline/
    securityhub_org/
    inspector_org/
    config_aggregation_role/
    config_aggregator/
    account_placement/
docs/
  01-canonical-event-model.md
  02-lifecycle-model.md
  03-compliance-alignment.md
  04-hipaa-alignment.md
  05-data-lake-architecture.md
  06-developer-onboarding-guide.md
  07-producer-integration-guide.md
  08-consumer-integration-guide.md
```

---

## Core Concepts

### **Canonical Event Model**
All events follow a strict envelope + payload structure.

### **Compliance Engine**
Every event is evaluated against rules stored in DynamoDB, cached in Redis, and persisted in RDS.

### **Data Lake Layers**
- Raw (immutable ingestion backup)  
- Golden (canonical, compliance‑annotated)  
- Curated (Glue‑generated datasets)  
- Archive (long‑term, KMS‑encrypted)  

### **Replay**
Deterministic reconstruction from Raw/Golden with compliance re‑evaluation.

---

## Security & Compliance

- KMS encryption everywhere  
- Deny unencrypted uploads  
- Public access blocked  
- Org‑level CloudTrail + Config  
- Centralized log archive bucket  
- Compliance metadata on every event  

---

## Deployment

### Org‑level baseline
```
cd infra/org-management
terraform init
terraform apply
```

### Environment (e.g., dev)
```
cd infra/envs/dev
terraform init
terraform apply
```

---

## Integration Guides

- Developer Onboarding  
- Producer Integration  
- Consumer Integration  
- Lifecycle Model  
- Compliance Alignment  
- HIPAA Alignment  
- Data Lake Architecture  

---

## Demo Flow

1. Send event → ALB → ECS  
2. Show Raw S3 object  
3. Show Golden S3 object  
4. Show compliance metadata  
5. Show Glue crawler + Athena query  
6. Show org‑level CloudTrail logs in Log Archive  

---

## Status

This repo represents a complete, interview‑ready, enterprise‑grade event streaming platform for healthcare workloads.
```
