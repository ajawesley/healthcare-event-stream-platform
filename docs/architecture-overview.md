# AcmeCo Event Stream Platform — Architecture Overview

## 1. Purpose

The AcmeCo Event Stream Platform (ESP) provides a secure, compliant, and scalable foundation for ingesting, processing, storing, and analyzing healthcare events. It is designed to:

- Normalize heterogeneous event payloads into canonical structures  
- Enforce compliance (HIPAA + internal policies) on every event  
- Provide governed, KMS‑encrypted storage across all layers  
- Support analytics, ML, and replay workflows  
- Offer clean integration surfaces for producers and consumers  

---

## 2. Core Components

### **Network & Security**
- VPC with public, private, and isolated subnets  
- ALB in public subnets  
- ECS ingestion service in private subnets  
- RDS PostgreSQL + Redis in isolated subnets  
- VPC endpoints for S3, DynamoDB, STS, Logs, etc.  
- Security groups enforcing least‑privilege connectivity  

### **Ingestion Pipeline**
- ALB → ECS ingestion service  
- Envelope + payload validation  
- Canonicalization inside ECS  
- Compliance evaluation inside ECS  
- Canonical JSON (with compliance metadata) written to Raw S3  
- Lambda triggers Glue job  
- Glue converts canonical JSON → canonical Parquet in Golden S3  

### **Compliance Engine**
- DynamoDB: rule definitions  
- Redis: rule cache  
- RDS: compliance metadata store (rules + evaluation state)  

### **Data Lake**
- Raw S3 bucket (canonical JSON + compliance metadata, immutable, KMS‑encrypted)  
- Golden S3 bucket (canonical Parquet, KMS‑encrypted)  
- Curated datasets via Glue  
- Glue Crawlers + Data Catalog  
- Athena for analytics  

### **Replay**
- Deterministic reconstruction from Raw canonical JSON  
- Glue reprocesses historical partitions into Golden  
- Compliance can be re‑evaluated independently if rules change  

### **Org‑Level Governance**
- AWS Organizations OU structure  
- SCP baseline  
- Org‑level CloudTrail + Config  
- Centralized log archive bucket (KMS‑encrypted, deny unencrypted uploads)  
- Security Hub + Inspector (detective + vulnerability scanning)  

---

## 3. Event Lifecycle

1. Producer sends event → ALB → ECS ingestion service  
2. ECS validates envelope + payload  
3. ECS canonicalizes HL7/FHIR/X12/Rest into canonical JSON  
4. ECS evaluates compliance rules (DynamoDB + Redis + RDS)  
5. ECS attaches compliance metadata to the canonical JSON  
6. ECS writes canonical JSON (with compliance metadata) to Raw S3  
7. S3 event triggers Lambda → Lambda starts Glue job  
8. Glue reads canonical JSON from Raw S3 and writes canonical Parquet to Golden S3 (compliance fields ignored)  
9. Glue crawlers update schemas for analytics and ML  
10. Consumers read canonical (Golden) or curated datasets  
11. Replay reprocesses Raw canonical JSON through Glue to regenerate Golden Parquet  

---

## 4. Security & Compliance

- KMS encryption everywhere (S3, RDS, DynamoDB, CloudTrail, Config)  
- Deny unencrypted uploads on all buckets  
- Public access blocked at bucket + org levels  
- Compliance metadata attached during ingestion  
- Full auditability via CloudTrail, Config, and S3 access logs  

---

## 5. Integration Surfaces

### Producers
- Envelope + payload model  
- Schema registry  
- PHI encryption  
- Classification rules  

### Consumers
- Idempotent processing  
- Compliance metadata usage  
- Replay‑safe behavior  

---

## 6. Summary

The AcmeCo ESP is a healthcare‑grade event platform that provides:

- Secure ingestion  
- Canonicalization  
- Compliance enforcement  
- Governed storage  
- Analytics + ML readiness  
- Replay capabilities  
- Org‑level governance  

It is designed to be enterprise‑ready, interview‑ready, and production‑aligned.
