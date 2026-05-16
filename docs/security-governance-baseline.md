# AcmeCo Event Stream Platform — Security & Governance Baseline

**Platform:** AcmeCo Event Stream Platform (ESP)  
**Audience:** Internal AcmeCo developers, platform engineers, security architects  
**Status:** Canonical reference — do not modify without Platform Architecture review  
**Last reviewed:** 2026‑05

---

## Overview

The AcmeCo ESP Security & Governance Baseline defines the mandatory controls, guardrails, and enforcement mechanisms that apply to all workloads running on the Event Stream Platform. These controls ensure that every event, service, and data store meets AcmeCo’s enterprise security requirements, HIPAA safeguards, and internal governance policies.

This baseline applies uniformly across:

- ingestion services  
- compliance engine  
- canonicalization pipeline  
- data lake landing zones  
- replay and archive systems  
- producer and consumer integrations  

The baseline is enforced through a combination of AWS‑native controls, Terraform modules, SCPs, IAM permission boundaries, and automated compliance checks.

---

## Governance Principles

AcmeCo ESP governance is built on six core principles:

1. **Least Privilege** — Every identity, service, and role receives only the permissions required to perform its function.  
2. **Separation of Duties** — Producers, consumers, compliance services, and platform operators have isolated access scopes.  
3. **Defense in Depth** — Multiple layers of controls protect PHI/PII across the entire lifecycle.  
4. **Immutable Auditability** — All access, changes, and rule evaluations are logged and retained for 7 years.  
5. **Secure by Default** — Encryption, classification, retention, and compliance evaluation are mandatory.  
6. **Continuous Compliance** — Automated checks ensure drift detection, misconfiguration detection, and policy enforcement.

---

## Security Architecture Components

The baseline includes the following mandatory components:

### **1. AWS Organizations & SCPs**
Service Control Policies enforce global guardrails:

- prohibit disabling CloudTrail  
- prohibit deleting log archive buckets  
- prohibit creating unencrypted S3 buckets  
- prohibit public S3 access  
- prohibit IAM wildcard permissions  
- prohibit creation of IAM users  
- restrict regions to approved list  
- restrict KMS key deletion  
- restrict network egress to approved endpoints  

SCPs apply to all AcmeCo ESP accounts: **ingestion**, **data**, **compliance**, **archive**, **shared services**.

---

### **2. IAM Permission Boundaries**
All IAM roles created by Terraform must attach the AcmeCo ESP permission boundary, which enforces:

- no privilege escalation  
- no ability to modify SCPs  
- no ability to modify permission boundaries  
- no ability to read/write PHI unless explicitly granted  
- no ability to create network paths outside approved VPCs  
- no ability to disable encryption  

This ensures even misconfigured roles cannot violate governance.

---

### **3. Centralized Logging & Audit**
All logs flow to the **Log Archive Account**, including:

- CloudTrail (organization‑wide)  
- VPC Flow Logs  
- ALB access logs  
- ECS task logs  
- Compliance Engine logs  
- Schema Registry access logs  
- Replay API logs  
- Archive access logs  

Logs are:

- immutable  
- encrypted  
- retained for 7 years  
- accessible only to Security & Compliance teams  

---

### **4. AWS Config & Continuous Compliance**
AWS Config enforces configuration rules such as:

- S3 buckets must be encrypted  
- RDS must be encrypted  
- DynamoDB must be encrypted  
- VPC Flow Logs must be enabled  
- Security groups must not allow 0.0.0.0/0 except on port 443  
- IAM roles must use permission boundaries  
- CloudTrail must be enabled  
- KMS keys must have rotation enabled  

Noncompliant resources trigger alerts and automated remediation.

---

### **5. Security Hub & GuardDuty**
Security Hub provides consolidated security posture visibility across all AcmeCo ESP accounts.

GuardDuty monitors:

- anomalous API calls  
- credential compromise  
- suspicious network activity  
- exfiltration attempts  
- reconnaissance behavior  

Findings are routed to the Security Operations Center (SOC) for triage.

---

### **6. Network Isolation**
AcmeCo ESP enforces strict network segmentation:

- **Public subnets:** ALBs only  
- **Private subnets:** ECS tasks, compliance engine, internal services  
- **Isolated subnets:** RDS, DynamoDB endpoints, Redis, S3 endpoints  

No PHI flows through public networks.

All outbound traffic is restricted via:

- VPC endpoints  
- NAT gateways with egress controls  
- SCP restrictions  
- Firewall rules (optional future‑state)

---

### **7. Encryption Baseline**
All data is encrypted:

- **In transit:** TLS 1.2+ everywhere  
- **At rest:** AES‑256 via KMS CMKs  
- **Field‑level:** Producers encrypt PHI/PII fields in payloads  
- **Key rotation:** Enabled for all CMKs  

KMS keys are scoped per environment and per data domain.

---

### **8. Secrets Management**
All secrets are stored in AWS Secrets Manager:

- database credentials  
- API keys  
- service tokens  
- encryption keys (wrapped, not stored raw)  

Secrets are:

- rotated automatically  
- encrypted with KMS  
- never stored in environment variables or code  
- accessed only by IAM roles with explicit grants  

---

### **9. Data Governance & Retention**
Retention is enforced via:

- S3 lifecycle policies  
- Glacier archival  
- compliance‑driven retention overrides  
- 7‑year minimum for PHI/PII  
- cryptographic erasure at end‑of‑life  

Data lineage is preserved through:

- event metadata  
- compliance metadata  
- audit logs  
- schema registry versioning  

---

## Platform Guardrails

The following guardrails are mandatory for all AcmeCo ESP workloads:

### **Guardrail 1 — No Unencrypted Data Stores**
All S3, RDS, DynamoDB, Redis, and EBS volumes must be encrypted.

### **Guardrail 2 — No Public Access**
No public S3 buckets, public RDS instances, or public ECS tasks.

### **Guardrail 3 — No Inline Credentials**
Credentials must never appear in:

- code  
- Terraform variables  
- environment variables  

### **Guardrail 4 — No Wildcard IAM Permissions**
`"*"` is prohibited except in tightly scoped permission boundaries.

### **Guardrail 5 — No PHI in Logs**
Logs must not contain:

- PHI  
- PII  
- raw payloads  

Compliance Engine masks sensitive fields before logging.

### **Guardrail 6 — No Cross‑Tenant Data Leakage**
Tenant boundaries are enforced via:

- routing rules  
- ACLs  
- compliance rules  
- partitioning  

### **Guardrail 7 — No Bypass of Compliance Engine**
All events must pass through:

1. ingestion  
2. canonicalization  
3. compliance evaluation  

No direct writes to downstream systems.

---

## Security Responsibilities

### **Platform Responsibilities**
AcmeCo ESP is responsible for:

- encryption  
- routing  
- compliance evaluation  
- audit logging  
- retention enforcement  
- access control  
- network isolation  
- continuous compliance  
- incident detection  

### **Producer Responsibilities**
Producers must:

- classify events correctly  
- encrypt PHI/PII fields  
- populate required envelope fields  
- retry ingestion failures safely  
- avoid sending PHI in logs or metadata  

### **Consumer Responsibilities**
Consumers must:

- honor classification  
- store PHI only in approved systems  
- implement idempotent processing  
- restrict access to authorized users  
- avoid exfiltration of PHI  

---

## Governance Enforcement Pipeline

AcmeCo ESP enforces governance through:

1. **Terraform Module Guardrails**  
   - validated inputs  
   - enforced encryption  
   - enforced IAM boundaries  

2. **CI/CD Policy Checks**  
   - Checkov  
   - OPA policies  
   - schema validation  
   - compliance rule validation  

3. **Runtime Enforcement**  
   - SCPs  
   - IAM boundaries  
   - Config rules  
   - Security Hub findings  

4. **Audit & Monitoring**  
   - CloudTrail  
   - Flow logs  
   - Compliance logs  
   - Access logs  

---

## Summary

The AcmeCo ESP Security & Governance Baseline ensures:

- strong HIPAA‑aligned protections  
- consistent enforcement of enterprise security policies  
- immutable auditability  
- safe handling of PHI/PII  
- secure‑by‑default infrastructure  
- continuous compliance across all environments  

This baseline is mandatory for all services and integrations on the AcmeCo Event Stream Platform.
