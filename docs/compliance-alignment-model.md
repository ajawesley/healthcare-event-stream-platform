# AcmeCo Event Stream Platform — Compliance Alignment Model

**Platform:** AcmeCo Event Stream Platform (ESP)  
**Audience:** Internal AcmeCo developers and platform engineers  
**Status:** Canonical reference — do not modify without Platform Architecture review  
**Last reviewed:** 2026‑05

---

## Overview

The Compliance Alignment Model defines how compliance rules are evaluated, applied, and propagated across the event lifecycle within the AcmeCo Event Stream Platform. Compliance is a first‑class platform concern: every event is classified, evaluated, and annotated with compliance metadata to ensure HIPAA, SOC2, and internal governance requirements are met.

Compliance alignment ensures:

- PHI/PII is handled safely and consistently  
- retention policies are enforced  
- access boundaries are respected  
- auditability is guaranteed  
- downstream systems receive explicit compliance outcomes  

This document describes the compliance engine, rule model, evaluation flow, and the compliance metadata added to canonical events.

---

## Compliance Engine Responsibilities

The Compliance Engine performs four core functions:

1. **Classification Enforcement**  
   Ensures the event’s declared `data_classification` is valid and consistent with payload content.

2. **Rule Evaluation**  
   Applies compliance rules based on event type, tenant, payload content, and metadata.

3. **Metadata Annotation**  
   Adds compliance metadata fields to the canonical event.

4. **Retention & Access Alignment**  
   Ensures retention policies and access boundaries align with classification and rule outcomes.

The engine operates synchronously during canonicalization, before the event is routed or consumed.

---

## Compliance Rule Model

Compliance rules are stored in DynamoDB and cached in Redis for low‑latency evaluation. Each rule has:

| Field | Description |
|---|---|
| `rule_id` | Unique identifier for auditability and lineage. |
| `rule_type` | Category of rule (e.g., `phi_masking`, `retention_override`, `tenant_restriction`). |
| `applies_to_event_type` | Event types or patterns the rule applies to. |
| `applies_to_tenant` | Optional tenant scoping. |
| `condition` | Boolean expression evaluated against the canonical event. |
| `action` | What to do when the rule fires (flag, override retention, mask fields, etc.). |
| `severity` | Informational, warning, or blocking. |

Rules are versioned and auditable. Changes require Platform Architecture approval.

---

## Compliance Evaluation Flow

Compliance evaluation occurs immediately after canonicalization and before routing.

```
RAW EVENT
   ↓
Canonicalization
   ↓
Compliance Evaluation
   ↓
Compliance Metadata Applied
   ↓
Routed to Stream
```

### Step 1 — Load Rules  
Rules are loaded from Redis (or DynamoDB fallback).

### Step 2 — Evaluate Conditions  
Each rule’s condition is evaluated against:

- envelope fields  
- canonical structures  
- payload fields  
- metadata  

### Step 3 — Apply Actions  
Actions may include:

- setting `compliance_flag = true`  
- setting `compliance_reason`  
- overriding `retention_policy`  
- masking or removing sensitive fields  
- blocking the event (rare; only for severe violations)

### Step 4 — Annotate Metadata  
The engine writes the compliance metadata fields onto the canonical event.

---

## Compliance Metadata Fields

These fields are added to every canonical event, regardless of whether a rule fired.

| Field | Type | Description |
|---|---|---|
| `compliance_applied` | bool | Always `true` after evaluation. |
| `compliance_flag` | bool | `true` if any rule fired. |
| `compliance_reason` | string | Human‑readable explanation of the rule outcome. |
| `compliance_rule_type` | string | Category of the rule that fired. |
| `compliance_rule_id` | string | Unique identifier of the rule that fired. |
| `compliance_timestamp` | string (ISO 8601 UTC) | When evaluation occurred. |

If no rule fires:

- `compliance_flag = false`  
- `compliance_reason = ""`  
- `compliance_rule_type = ""`  
- `compliance_rule_id = ""`  

---

## Retention Alignment

Retention policies must align with classification and rule outcomes.

### Default Retention Rules

| Classification | Default Retention |
|---|---|
| `phi` | 7 years |
| `pii` | 7 years |
| `internal` | 3 years |
| `public` | 90 days |

### Rule‑Driven Overrides

Rules may override retention, for example:

- extend retention for audit‑critical events  
- shorten retention for ephemeral operational events  
- enforce stricter retention for specific tenants  

Overrides are recorded in `compliance_reason`.

---

## Access Control Alignment

Compliance evaluation ensures that:

- PHI/PII events are routed only to streams with PHI‑approved ACLs  
- consumers without PHI access cannot subscribe to PHI topics  
- tenant‑restricted events cannot cross tenant boundaries  
- masking rules are applied before routing  

If a rule determines an event must not be routed to a consumer group, the platform enforces this at the routing layer.

---

## Classification Enforcement

The Compliance Engine validates that:

- `data_classification` is present and valid  
- payload content is consistent with classification  
- producers are not misclassifying PHI as `internal` or `public`  

If misclassification is detected:

- `compliance_flag = true`  
- `compliance_reason = "classification_mismatch"`  
- retention is automatically elevated to `7y`  
- routing is restricted to PHI‑approved streams  

---

## Example Compliance Outcomes

### Example 1 — No Rules Fired

```
"compliance_applied": true,
"compliance_flag": false,
"compliance_reason": "",
"compliance_rule_type": "",
"compliance_rule_id": "",
"compliance_timestamp": "2026-05-01T14:23:01Z"
```

### Example 2 — Retention Override Rule Fired

```
"compliance_applied": true,
"compliance_flag": true,
"compliance_reason": "retention_override: audit_critical",
"compliance_rule_type": "retention_override",
"compliance_rule_id": "RET-2026-0041",
"compliance_timestamp": "2026-05-01T14:23:01Z"
```

### Example 3 — PHI Masking Rule Fired

```
"compliance_applied": true,
"compliance_flag": true,
"compliance_reason": "phi_masking: sensitive_lab_value",
"compliance_rule_type": "phi_masking",
"compliance_rule_id": "PHI-2026-0098",
"compliance_timestamp": "2026-05-01T14:23:01Z"
```

---

## Auditability & Lineage

Compliance evaluation is fully auditable:

- every rule evaluation is logged  
- every rule change is versioned  
- every compliance outcome is attached to the canonical event  
- lineage records include:  
  - rule ID  
  - rule version  
  - evaluation timestamp  
  - event ID  

Audit logs are immutable and retained for 7 years.

---

## Failure Modes

| Failure | Platform Behavior | Notes |
|---|---|---|
| Rule store unavailable | Fallback to cached rules | Platform never ingests without rules |
| Rule evaluation error | Event flagged, not rejected | `compliance_flag = true` |
| Severe violation | Event rejected | Rare; requires explicit rule configuration |
| Misclassification | Auto‑elevate retention + restrict routing | Logged for audit |

---

## Summary

The Compliance Alignment Model ensures:

- consistent enforcement of PHI/PII rules  
- explicit compliance metadata on every event  
- safe routing and retention  
- auditability and lineage  
- predictable behavior for downstream consumers  

Compliance is not optional — it is a core platform guarantee.
