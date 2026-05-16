```md
# AcmeCo Event Stream Platform (ESP)

A secure, compliant, healthcare‑grade event streaming and data platform designed to:

- Ingest, validate, and normalize heterogeneous events  
- Canonicalize payloads into a unified event model  
- Enforce HIPAA and internal compliance policies  
- Store data in a governed, KMS‑encrypted data lake  
- Support analytics, ML, replay, and lineage workflows  
- Provide clean, stable integration surfaces for producers and consumers  

---

## 📁 Repository Structure

```
infra/
  org-management/                # Org-level: OUs, SCPs, CloudTrail, Config, log archive
  envs/
    dev/
      main.tf                    # Dev environment: VPC, ECS, RDS, Redis, S3, Glue, Config
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

## 🧠 Core Concepts

### **Canonical Event Model**
All events follow a strict **envelope + payload** structure to ensure consistency across FHIR, HL7, X12, and JSON producers.

### **Compliance Engine**
Every event is evaluated against compliance rules stored in DynamoDB, cached in Redis, and persisted in RDS for auditability.

### **Data Lake Architecture**
The platform uses a layered, governed data lake:

- **Raw** — Immutable ingestion backup  
- **Golden** — Canonical, compliance‑annotated events  
- **Curated** — Glue‑generated analytics datasets  
- **Archive** — Long‑term, KMS‑encrypted storage  

### **Replay**
Deterministic reconstruction of events from Raw/Golden with full compliance re‑evaluation.

---

## 🔐 Security & Compliance

The platform enforces strict enterprise‑grade controls:

- KMS encryption on all data at rest  
- Deny unencrypted uploads  
- Block all public access  
- Organization‑level CloudTrail + Config  
- Centralized log archive bucket  
- Compliance metadata attached to every event  
- Deterministic lineage for audit and replay  

---

## 🚀 Deployment

### **1. Organization‑Level Baseline**
```
cd infra/org-management
terraform init
terraform apply
```

### **2. Environment Deployment (e.g., dev)**
```
cd infra/envs/dev
terraform init
terraform apply
```

---

## 📘 Integration Guides

Documentation for all platform users:

- **Developer Onboarding Guide**  
- **Producer Integration Guide**  
- **Consumer Integration Guide**  
- **Lifecycle Model**  
- **Compliance Alignment Model**  
- **HIPAA Alignment**  
- **Data Lake Architecture**  

All located in the `docs/` directory.

---

## 🎬 Demo Flow (Recommended Walkthrough)

1. Send event → ALB → ECS ingestion service  
2. Inspect Raw S3 object  
3. Inspect Golden S3 object  
4. Review compliance metadata  
5. Run Glue crawler + Athena query  
6. View org‑level CloudTrail logs in Log Archive  

This demonstrates ingestion, canonicalization, compliance, lineage, and analytics readiness.

---

## 📌 Status

This repository represents a complete, demonstrable, enterprise‑grade event streaming platform tailored for healthcare workloads, with compliance, governance, and observability built in from day one.
```
