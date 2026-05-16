# AcmeCo Event Stream Platform — Event Lifecycle Model

**Platform:** AcmeCo Event Stream Platform (ESP)  
**Audience:** Internal AcmeCo developers and platform engineers  
**Status:** Canonical reference — do not modify without Platform Architecture review  
**Last reviewed:** 2026‑05

---

## Overview

The Event Lifecycle Model defines how events move through the AcmeCo Event Stream Platform from initial production to archival and replay. The lifecycle ensures:

- consistent ingestion  
- deterministic canonicalization  
- compliance evaluation  
- secure routing  
- governed retention  
- auditable replay  

This document describes each lifecycle stage, the platform guarantees, and the responsibilities of producers and consumers.

---

## Lifecycle Stages

The lifecycle of an event in the AcmeCo ESP consists of the following stages:

```
PRODUCED → INGESTED → CANONICALIZED → COMPLIANCE EVALUATED → ROUTED → CONSUMED → RETAINED → ARCHIVED → REPLAYED (optional)
```

Each stage is governed by strict platform rules to ensure safety, compliance, and auditability.

---

## Stage 1 — Produced

A producer constructs an event containing:

- envelope fields  
- payload fields  
- field‑level encrypted PHI/PII (if applicable)  

**Platform guarantees:**

- none — the event has not yet entered the platform  

**Producer responsibilities:**

- generate UUID v4 `event_id`  
- set `produced_at` in UTC  
- classify data correctly (`phi`, `pii`, `internal`, `public`)  
- encrypt PHI/PII fields  
- validate payload schema  
- register `source_system`  

---

## Stage 2 — Ingested

The event is submitted to the ESP ingestion API via mTLS.

**Platform actions:**

- validate envelope  
- set `ingested_at`  
- reject malformed events  
- write raw event to durable storage (S3)  

**Platform guarantees:**

- at‑least‑once ingestion  
- durable persistence  
- no data loss after acceptance  

---

## Stage 3 — Canonicalized

The transformation layer normalizes the event into canonical structures:

- `patient`  
- `encounter`  
- `observation`  
- `metadata`  

**Platform guarantees:**

- deterministic canonicalization  
- schema‑driven transformation  
- lineage preservation  

---

## Stage 4 — Compliance Evaluated

The Compliance Engine evaluates the event against AcmeCo compliance rules.

**Platform actions:**

- load rules from cache  
- evaluate rule conditions  
- apply masking or retention overrides  
- annotate compliance metadata  

**Platform guarantees:**

- every event receives compliance metadata  
- misclassification is corrected  
- PHI/PII routing restrictions enforced  

---

## Stage 5 — Routed

The event is routed to the appropriate stream based on:

- `event_type`  
- `tenant_id`  
- `data_classification`  
- compliance outcomes  

**Platform guarantees:**

- PHI/PII only flows to approved streams  
- tenant isolation  
- partition‑level ordering  

---

## Stage 6 — Consumed

Downstream services read events from their authorized streams.

**Platform guarantees:**

- at‑least‑once delivery  
- partition‑level ordering  
- offset tracking  

**Consumer responsibilities:**

- implement idempotency  
- honor classification  
- store PHI only in approved systems  
- handle replay safely  

---

## Stage 7 — Retained

Events remain on the active stream for a minimum of **90 days**.

**Platform guarantees:**

- 90‑day minimum retention  
- retention aligned with classification and compliance rules  

---

## Stage 8 — Archived

After active retention expires, events move to cold storage (Glacier).

**Platform guarantees:**

- 1–7 year archival depending on classification  
- immutable audit logs  
- cryptographic erasure at end‑of‑life  

---

## Stage 9 — Replayed (Optional)

Historical events may be reprocessed through the Replay subsystem.

**Platform guarantees:**

- deterministic reconstruction  
- compliance re‑evaluation  
- isolation from real‑time ingestion  
- full auditability  

Replay is governed by strict access controls and approval workflows.

---

## Platform Guarantees Summary

| Stage | Guarantee |
|---|---|
| Produced | none |
| Ingested | durable storage, at‑least‑once ingestion |
| Canonicalized | deterministic transformation |
| Compliance | rule evaluation + metadata |
| Routed | PHI/PII isolation, tenant isolation |
| Consumed | at‑least‑once delivery, ordering within partition |
| Retained | 90‑day minimum |
| Archived | 1–7 year retention |
| Replayed | deterministic, auditable, isolated |

---

## Responsibilities Summary

### Producers Must:

- classify data correctly  
- encrypt PHI/PII  
- validate payload schema  
- generate UUID v4 `event_id`  
- set `produced_at`  
- avoid setting `ingested_at`  

### Consumers Must:

- implement idempotency  
- respect compliance metadata  
- store PHI only in approved systems  
- handle replay safely  

### Platform Provides:

- ingestion  
- canonicalization  
- compliance evaluation  
- routing  
- retention  
- archival  
- replay  
- audit logging  

---

## Summary

The AcmeCo Event Lifecycle Model ensures that every event is:

- safely ingested  
- normalized  
- compliance‑aligned  
- securely routed  
- retained and archived  
- replayable with full auditability  

This lifecycle is foundational to the reliability, security, and compliance posture of the AcmeCo Event Stream Platform.
