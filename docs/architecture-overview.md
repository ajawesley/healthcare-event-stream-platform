```md
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
- Raw event storage (S3)  
- Canonicalization into Golden S3  
- Compliance evaluation using DynamoDB + Redis + RDS  

### **Compliance Engine**
- DynamoDB: rule definitions  
- Redis: rule cache  
- RDS: compliance metadata, lineage, audit trail  

### **Data Lake**
- Raw S3 bucket (immutable, KMS‑encrypted)  
- Golden S3 bucket (canonical, KMS‑encrypted)  
- Curated datasets via Glue  
- Glue Crawlers + Data Catalog  
- Athena for analytics  

### **Replay**
- Deterministic reconstruction from Raw/Golden  
- Re‑evaluation of compliance  
- Re‑emission into replay topics or downstream systems  

### **Org‑Level Governance**
- AWS Organizations OU structure  
- SCP baseline  
- Org‑level CloudTrail + Config  
- Centralized log archive bucket (KMS‑encrypted, deny unencrypted uploads)  
- Security Hub + Inspector  

---

## 3. Event Lifecycle

1. Producer sends event → ALB → ECS  
2. ECS validates envelope + payload  
3. Raw event written to Raw S3  
4. Canonical event written to Golden S3  
5. Compliance engine evaluates rules  
6. Compliance metadata attached  
7. Glue crawlers update schemas  
8. Consumers read canonical or curated datasets  
9. Replay reconstructs events from Raw/Golden  

---

## 4. Security & Compliance

- KMS encryption everywhere (S3, RDS, DynamoDB, CloudTrail, Config)  
- Deny unencrypted uploads on all buckets  
- Public access blocked at bucket + org levels  
- Compliance metadata attached to every event  
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
```
