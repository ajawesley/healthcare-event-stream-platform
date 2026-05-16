# AcmeCo Event Stream Platform — Replay Model

**Platform:** AcmeCo Event Stream Platform (ESP)  
**Audience:** Internal platform engineers, service teams, compliance teams  
**Status:** Canonical reference — do not modify without Platform Architecture review  
**Last reviewed:** 2026‑05

---

## Overview

Replay is a controlled, auditable mechanism that allows consumers and platform services to reprocess historical events. Replay is essential for:

- state reconstruction  
- backfills  
- compliance re-evaluation  
- downstream system recovery  
- analytics corrections  
- incident response  

Replay is **not** a simple re-delivery of events. It is a governed workflow with strict access controls, auditability, and compliance alignment.

Replay operates on events stored in:

- the **active stream** (90‑day minimum retention)  
- the **cold archive** (Glacier, 1–7 years depending on classification)  

This document defines the replay model, access controls, lifecycle, and platform guarantees.

---

## Replay Principles

Replay is governed by five core principles:

1. **Safety First** — Replay must never corrupt downstream state or violate PHI/PII boundaries.  
2. **Auditability** — Every replay request, execution, and outcome is logged.  
3. **Determinism** — Replayed events must be identical to their original canonical form.  
4. **Isolation** — Replay must not interfere with real‑time ingestion or consumption.  
5. **Least Privilege** — Only authorized teams may initiate replay, and only for approved tenants and event types.

---

## Replay Sources

Replay can pull events from two storage layers:

### **1. Active Stream (0–90 days)**  
- Low latency  
- Ordered by partition  
- Ideal for consumer offset resets  
- No retrieval delay  

### **2. Cold Archive (90 days–7 years)**  
- Glacier Deep Archive  
- Higher latency (minutes to hours)  
- Requires explicit approval  
- Ideal for compliance investigations, audits, and backfills  

Replay from archive is always slower and more tightly governed.

---

## Replay Types

AcmeCo ESP supports three replay modes.

---

### **1. Offset Reset Replay (Consumer‑Driven)**  
Consumers may reset their committed offset to any earlier point within the active stream retention window.

**Use cases:**

- consumer bug fix  
- downstream outage recovery  
- reprocessing due to schema change  

**Characteristics:**

- no platform operator involvement  
- no event mutation  
- no compliance re-evaluation  
- ordering preserved within partition  

---

### **2. Targeted Replay (Platform‑Driven)**  
Replay a specific subset of events based on:

- event_id  
- correlation_id  
- event_type  
- tenant_id  
- time window  

**Use cases:**

- reconstruct a specific business transaction  
- reprocess a failed batch  
- recover from partial ingestion failures  
- compliance re-evaluation for a subset of events  

**Characteristics:**

- requires replay request approval  
- events are re-canonicalized  
- compliance rules are re-evaluated  
- routing follows current rules, not historical ones  

---

### **3. Full Backfill Replay (Bulk Replay)**  
Replay a large historical dataset from archive.

**Use cases:**

- new downstream system onboarding  
- analytics backfill  
- compliance-driven reprocessing  
- migration to new canonical schema version  

**Characteristics:**

- requires formal change request  
- may take hours or days  
- executed in controlled batches  
- monitored by platform operators  
- strict rate limiting to avoid downstream overload  

---

## Replay Lifecycle

Replay follows a structured lifecycle:

```
REQUESTED → APPROVED → PREPARED → EXECUTING → COMPLETED
                    ↓
                 REJECTED (terminal)
```

---

### Stage 1 — REQUESTED

A replay request is submitted via:

- Replay API  
- Platform Console  
- Internal ticket (for archive replay)  

**Required fields:**

- requesting team  
- tenant_id  
- event_type(s)  
- replay type  
- time window or identifiers  
- justification  
- downstream impact assessment  

---

### Stage 2 — APPROVED

Replay requires approval from:

- Platform Operations  
- Compliance  
- Data Governance (for large backfills)  

Approval ensures:

- replay is safe  
- replay is necessary  
- replay will not violate PHI/PII boundaries  
- downstream systems are prepared  

---

### Stage 3 — PREPARED

Platform prepares the replay job:

- identifies source storage (stream or archive)  
- retrieves metadata  
- validates retention windows  
- prefetches archive objects (if needed)  
- allocates replay workers  
- applies rate limits  

---

### Stage 4 — EXECUTING

Replay workers:

1. Retrieve events  
2. Reconstruct canonical form  
3. Re-evaluate compliance rules  
4. Re-emit events into replay topics  
5. Preserve ordering within partitions  
6. Log every replayed event  

Replay topics are isolated from real‑time ingestion topics.

---

### Stage 5 — COMPLETED

Replay job is marked complete when:

- all events have been reprocessed  
- all compliance evaluations succeeded  
- all downstream acknowledgements are received (optional)  
- audit logs are finalized  

A replay summary is generated:

- number of events replayed  
- time window  
- rule outcomes  
- downstream consumers impacted  
- any failures or retries  

---

### Terminal State — REJECTED

Replay is rejected if:

- request lacks justification  
- request violates PHI/PII boundaries  
- time window exceeds retention  
- requester lacks authorization  
- downstream systems are not prepared  

Rejection is logged and visible to the requester.

---

## Replay Guarantees

| Guarantee | Scope | Notes |
|---|---|---|
| Deterministic event reconstruction | All replay types | Canonical form is identical to original |
| Ordering | Within partition | Same as ingestion |
| At-least-once replay | All replay types | Consumers must dedupe |
| Compliance re-evaluation | Targeted & full replay | Ensures alignment with current rules |
| Isolation from real-time traffic | All replay types | Replay uses separate topics |
| Auditability | All replay types | Every replayed event is logged |

Replay **does not** guarantee:

- cross-partition ordering  
- exactly-once delivery  
- preservation of historical routing rules  

---

## Compliance Alignment During Replay

Replay always re-evaluates compliance rules for:

- targeted replay  
- full backfill replay  
- archive replay  

This ensures:

- updated PHI masking rules apply  
- updated retention policies apply  
- updated tenant restrictions apply  
- misclassified historical events are corrected  

Compliance metadata is regenerated with a new `compliance_timestamp`.

---

## Replay Access Control

Replay is a privileged operation.

### **Allowed to request replay:**

- platform operators  
- compliance teams  
- data governance  
- approved consumer teams (offset reset only)  

### **Not allowed to request replay:**

- producers  
- external vendors  
- unapproved consumers  

Replay requests are logged and retained for 7 years.

---

## Failure Handling

| Failure | Platform Behavior | Notes |
|---|---|---|
| Archive retrieval failure | Retry with exponential backoff | Glacier delays are normal |
| Compliance rule error | Event flagged, not dropped | Logged for audit |
| Downstream overload | Replay throttled | Protects consumers |
| Replay worker failure | Automatic retry | Logged |
| Invalid request | Rejected | No partial replay |

---

## Summary

The AcmeCo Replay Model provides:

- safe, auditable reprocessing of historical events  
- deterministic reconstruction of canonical events  
- compliance re-evaluation  
- strict access controls  
- isolation from real-time ingestion  
- support for backfills, investigations, and recovery  

Replay is a powerful capability — and a tightly governed one — ensuring that historical data can be reprocessed without compromising security, compliance, or platform stability.

