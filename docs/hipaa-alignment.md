# AcmeCo Event Stream Platform — HIPAA Alignment Model

**Platform:** AcmeCo Event Stream Platform (ESP)  
**Audience:** Internal AcmeCo developers, platform engineers, compliance teams  
**Status:** Canonical reference — do not modify without Platform Architecture review  
**Last reviewed:** 2026‑05

---

## Overview

The AcmeCo Event Stream Platform (ESP) is designed to meet or exceed the administrative, technical, and physical safeguards required under the HIPAA Security Rule (45 CFR §164.308, §164.310, §164.312) and the HIPAA Privacy Rule (45 CFR §164.500–534).

This document describes how ESP aligns with HIPAA requirements across:

- data classification  
- encryption  
- access control  
- audit logging  
- retention  
- data minimization  
- breach detection and response  
- compliance metadata  
- producer and consumer responsibilities  

These protections apply automatically to all events processed through ESP.

---

## HIPAA Safeguard Mapping

ESP implements safeguards across all three HIPAA categories:

### **1. Administrative Safeguards (45 CFR §164.308)**  
- Access management via stream ACLs  
- Workforce access provisioning through ESP Access  
- Security incident procedures  
- Periodic risk assessments  
- Policy‑driven retention and archival  
- Producer onboarding and validation workflows  

### **2. Physical Safeguards (45 CFR §164.310)**  
- AWS data center protections (SOC2, ISO 27001, FedRAMP)  
- No PHI stored on developer workstations  
- No local caching of PHI in ESP services  
- All PHI stored only in encrypted AWS services  

### **3. Technical Safeguards (45 CFR §164.312)**  
- mTLS for ingestion  
- AES‑256 encryption at rest  
- IAM least privilege  
- PHI‑aware routing  
- Immutable audit logs  
- Automatic compliance metadata  
- Retention enforcement  
- Replay with auditability  

---

## Data Classification Alignment

Every event must declare a `data_classification` in the envelope:

- `phi`  
- `pii`  
- `internal`  
- `public`  

The Compliance Engine validates classification and elevates it if misclassified.

### **Classification Enforcement Guarantees**

| Classification | Encryption | Routing          | Retention | Access       |
|----------------|------------|------------------|-----------|--------------|
| `phi`          | Required   | PHI‑restricted   | 7 years   | Explicit ACL |
| `pii`          | Required   | PII‑restricted   | 7 years   | Explicit ACL |
| `internal`     | Required   | Internal streams | 3 years   | Internal ACL |
| `public`       | Optional   | Public streams   | 90 days   | Open         |

Misclassification triggers:

- `compliance_flag = true`  
- retention override to `7y`  
- routing restricted to PHI‑approved streams  

---

## Encryption Requirements

### **Encryption in Transit (45 CFR §164.312(e))**

All ingestion endpoints require:

- **mTLS (TLS 1.2+)**  
- client certificate validation  
- no plaintext HTTP  
- no weak cipher suites  

Internal service‑to‑service calls also use TLS.

### **Encryption at Rest (45 CFR §164.312(a)(2)(iv))**

All PHI/PII is encrypted using:

- **AES‑256**  
- **AWS KMS CMKs**  
- automatic key rotation  
- envelope encryption for S3, RDS, DynamoDB, Redis  

Producers must encrypt PHI/PII fields inside the payload before sending.

---

## Access Control Alignment

ESP enforces strict access boundaries:

### **Stream ACLs**
- Consumers must be explicitly granted access to PHI/PII streams.  
- ACLs are tenant‑scoped and event‑type‑scoped.  
- No wildcard access for PHI.

### **IAM Least Privilege**
- Execution roles have only the permissions required for ingestion, routing, and compliance.  
- Task roles have only the permissions required for reading/writing specific data stores.  

### **Separation of Duties**
- Producers cannot read streams.  
- Consumers cannot write to ingestion.  
- Compliance Engine cannot modify payloads except for masking rules.

---

## Audit Logging & Monitoring

ESP maintains immutable audit logs for:

- ingestion events  
- compliance rule evaluations  
- PHI/PII access  
- consumer reads  
- replay operations  
- retention and archival actions  
- rule changes  
- ACL changes  

Audit logs are:

- immutable  
- encrypted  
- retained for **7 years**  
- stored in the Log Archive account  
- accessible only to compliance and security teams  

This satisfies HIPAA §164.312(b) — Audit Controls.

---

## Retention & Archival Alignment

Retention is governed by the envelope’s `retention_policy`, with HIPAA‑aligned defaults:

| Classification | Minimum Retention | Archive Requirement |
|--- ------------|---------|----------|
| PHI            | 7 years | Required |
| PII            | 7 years | Required |
| Internal       | 3 years | Optional |
| Public         | 90 days | Optional |

### **Active Stream Retention**
All events remain on the active stream for **90 days minimum**, regardless of declared policy.

### **Cold Archive**
After stream retention expires:

- events move to Glacier Deep Archive  
- access requires explicit approval  
- retrieval is audited  
- deletion uses cryptographic erasure  

This satisfies HIPAA §164.530(j) — Documentation Retention.

---

## Data Minimization & Masking

ESP enforces:

- field‑level encryption by producers  
- optional masking rules via Compliance Engine  
- removal of unnecessary PHI fields  
- payload validation to prevent PHI in `internal` or `public` events  

Masking rules may:

- redact sensitive values  
- hash identifiers  
- remove unnecessary fields  
- override routing to PHI‑restricted streams  

---

## Breach Detection & Incident Response

ESP integrates with enterprise monitoring to detect:

- unauthorized access attempts  
- anomalous read patterns  
- misrouted PHI  
- consumer exfiltration attempts  
- rule evaluation failures  

When a potential breach is detected:

1. Access is automatically restricted.  
2. Compliance and Security teams are alerted.  
3. Audit logs are frozen for investigation.  
4. AcmeCo’s Incident Response Plan is initiated.  

This satisfies HIPAA §164.308(a)(6) — Security Incident Procedures.

---

## Compliance Metadata Alignment

Compliance metadata fields on canonical events provide:

- explicit rule outcomes  
- auditability  
- lineage  
- downstream enforcement signals  

These fields ensure downstream systems can:

- enforce PHI boundaries  
- apply retention correctly  
- mask or restrict data  
- log compliance outcomes  

This satisfies HIPAA’s requirement for **traceability** and **accountability**.

---

## Producer Responsibilities

Producers must:

- classify events correctly  
- encrypt PHI/PII fields before submission  
- populate all required envelope fields  
- avoid sending PHI in `internal` or `public` events  
- retry ingestion failures safely using `event_id`  

Producers **must not**:

- set `ingested_at`  
- bypass classification  
- send unencrypted PHI  
- embed PHI in metadata fields  

---

## Consumer Responsibilities

Consumers must:

- honor classification  
- store PHI only in approved systems  
- implement idempotent processing  
- avoid speculative acknowledgements  
- restrict access to authorized users  

Consumers **must not**:

- write PHI to non‑approved stores  
- exfiltrate PHI outside AcmeCo networks  
- bypass ACLs  
- ignore compliance metadata  

---

## Summary

The AcmeCo ESP provides a HIPAA‑aligned foundation for healthcare event ingestion and processing by enforcing:

- encryption  
- access control  
- retention  
- auditability  
- compliance metadata  
- classification enforcement  
- masking and minimization  
- incident detection  

These guarantees ensure that all events processed through ESP meet or exceed HIPAA requirements across their entire lifecycle.
