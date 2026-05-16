# AcmeCo Event Stream Platform — Data Lake Architecture

**Platform:** AcmeCo Event Stream Platform (ESP)  
**Audience:** Platform engineers, data engineering teams, analytics teams  
**Status:** Canonical reference — do not modify without Platform Architecture review  
**Last reviewed:** 2026‑05

---

## Overview

The AcmeCo Event Stream Platform (ESP) Data Lake provides a governed, secure, and analytics‑ready storage layer for all events processed through ESP. It supports:

- raw event retention  
- canonical event storage  
- compliance‑aligned archival  
- analytics and ML workloads  
- replay and lineage reconstruction  
- schema evolution and versioning  

The Data Lake is designed to meet strict requirements for PHI/PII protection, auditability, and long‑term retention.

---

## Data Lake Layers

The AcmeCo Data Lake is composed of four logical layers:

```
RAW → CANONICAL → CURATED → ARCHIVE
```

Each layer has a distinct purpose, retention policy, and access model.

---

## Layer 1 — RAW (Immutable Ingestion Backup)

**Purpose:**  
Durable, immutable backup of every event as originally ingested.

**Characteristics:**

- written synchronously by the ingestion API  
- contains the exact producer‑submitted envelope + payload  
- no transformations  
- no compliance metadata  
- immutable (write‑once)  
- encrypted with KMS  
- stored in S3 with Object Lock (governance mode)  

**Use Cases:**

- audit investigations  
- replay reconstruction  
- debugging ingestion issues  
- lineage verification  

**Retention:**  
90 days minimum (configurable per classification).

---

## Layer 2 — CANONICAL (Platform‑Normalized Events)

**Purpose:**  
Store the fully transformed, compliance‑evaluated canonical event.

**Characteristics:**

- includes canonical structures (`patient`, `encounter`, `observation`)  
- includes compliance metadata  
- includes lineage metadata  
- partitioned by:  
  ```
  /event_type=<type>/year=<yyyy>/month=<mm>/day=<dd>/
  ```  
- stored in Parquet format  
- schema‑versioned  
- optimized for analytics  

**Use Cases:**

- analytics  
- ML feature pipelines  
- compliance audits  
- replay source of truth  

**Retention:**  
Aligned with `retention_policy` (1–7 years).

---

## Layer 3 — CURATED (Domain‑Optimized Views)

**Purpose:**  
Provide domain‑specific, analytics‑ready datasets derived from canonical events.

**Characteristics:**

- created by data engineering teams  
- may join multiple canonical event types  
- may include derived fields  
- may include PHI masking or tokenization  
- stored in Parquet or Iceberg  
- versioned and documented  

**Use Cases:**

- BI dashboards  
- ML training datasets  
- operational analytics  
- domain‑specific reporting  

**Retention:**  
Domain‑specific; defaults to 3 years unless PHI requires 7.

---

## Layer 4 — ARCHIVE (Cold Storage)

**Purpose:**  
Long‑term, compliance‑aligned archival of canonical events.

**Characteristics:**

- stored in Glacier Deep Archive  
- retrieval latency: minutes to hours  
- immutable  
- encrypted  
- access requires approval  
- used for replay, audits, and investigations  

**Retention:**  
1–7 years depending on classification and compliance rules.

---

## Data Lake Storage Format

### **Primary Format: Parquet**

Chosen for:

- columnar compression  
- schema evolution support  
- efficient analytics  
- compatibility with Spark, Athena, EMR, Glue  

### **Optional Format: Iceberg**

Used for curated datasets requiring:

- ACID transactions  
- time travel  
- schema evolution  
- partition pruning  

---

## Partitioning Strategy

Canonical events are partitioned by:

```
event_type=<type>/year=<yyyy>/month=<mm>/day=<dd>/
```

This provides:

- efficient filtering  
- predictable file layout  
- compatibility with Athena and Spark  
- replay‑friendly organization  

---

## Schema Evolution

The Data Lake supports schema evolution through:

- semantic versioning (`event_version`)  
- Parquet schema evolution  
- Iceberg schema evolution (for curated datasets)  
- schema registry enforcement  

### **Backward‑Compatible Changes (Allowed)**

- adding optional fields  
- widening field types  
- adding new canonical structures  

### **Breaking Changes (Require Major Version)**

- removing fields  
- renaming fields  
- changing field types incompatibly  

---

## Compliance Alignment

The Data Lake enforces compliance through:

### **1. Classification‑Aligned Retention**
Retention is driven by:

- `data_classification`  
- `retention_policy`  
- compliance rule overrides  

### **2. PHI/PII Isolation**
PHI/PII datasets:

- stored in PHI‑restricted buckets  
- require explicit ACLs  
- are encrypted with PHI‑scoped KMS keys  

### **3. Auditability**
All access to:

- RAW  
- CANONICAL  
- ARCHIVE  

is logged and retained for 7 years.

### **4. Masking & Tokenization**
Curated datasets may apply:

- hashing  
- tokenization  
- redaction  
- field removal  

---

## Replay Integration

Replay uses the Data Lake as a source of truth:

- RAW layer for exact producer input  
- CANONICAL layer for deterministic reconstruction  
- ARCHIVE layer for long‑term replay  

Replay workers read from the Data Lake, reconstruct canonical events, re‑evaluate compliance, and re‑emit events into replay topics.

---

## Access Control Model

### **RAW Layer**
- restricted to platform operators  
- no consumer access  

### **CANONICAL Layer**
- accessible to analytics teams  
- PHI requires explicit approval  

### **CURATED Layer**
- domain‑specific access  
- PHI masking may be applied  

### **ARCHIVE Layer**
- access requires compliance approval  
- retrieval is fully audited  

---

## Data Lake Governance

Governance is enforced through:

- S3 bucket policies  
- IAM permission boundaries  
- SCPs  
- AWS Config rules  
- Glue Data Catalog schema validation  
- automated retention policies  
- audit logging  

---

## Summary

The AcmeCo ESP Data Lake provides:

- immutable raw event storage  
- normalized canonical datasets  
- curated analytics‑ready views  
- long‑term archival  
- compliance‑aligned retention  
- PHI/PII protection  
- replay support  
- schema evolution  

This architecture ensures that all event data is secure, auditable, analytics‑ready, and aligned with AcmeCo’s compliance and governance requirements.
