```md
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

# Compliance Rules Reference (Authoritative Rule Catalog)

This section provides the canonical reference for all compliance rules defined in the AcmeCo Event Stream Platform. These rules represent the business logic used to evaluate FHIR, Generic JSON, X12, and HL7 events during canonicalization and compliance evaluation.

Each rule includes:
- Business purpose  
- Trigger conditions  
- Applicable event formats  
- Impact on compliance metadata  
- Impact on routing, retention, or masking  

---

## Rule Categories

| Category | Description |
|---|---|
| **PHI Masking** | Detects sensitive clinical or demographic data and masks or removes fields. |
| **Retention Override** | Extends or shortens retention based on regulatory or audit requirements. |
| **Tenant Restriction** | Ensures events cannot cross tenant boundaries. |
| **Classification Enforcement** | Ensures declared classification matches actual payload content. |

---

# Rule Catalog With Business Descriptions

## PHI‑001 — Sensitive Lab Value Masking
**Category:** `phi_masking`  
**Business Purpose:** Mask specific lab values considered high‑risk PHI (e.g., HIV markers, genetic tests).

**Triggers:**  
- Sensitive lab codes present  
- Event classification is `phi` or `pii`

**Applies To:**  
- **FHIR:** Observation, DiagnosticReport  
- **HL7:** OBX  
- **X12:** Clinical attachment loops  
- **Generic JSON:** `lab_code`, `lab_value`

**Actions:**  
- Mask fields  
- `compliance_flag = true`  
- `compliance_rule_type = "phi_masking"`

---

## PHI‑002 — Demographic Sensitivity
**Category:** `phi_masking`  
**Business Purpose:** Mask demographic fields that increase re‑identification risk.

**Triggers:**  
- Presence of demographic fields  
- Classification is `phi` or `pii`

**Applies To:**  
- **FHIR:** Patient, Encounter  
- **HL7:** PID  
- **X12:** Subscriber loops  
- **Generic JSON:** `address`, `phone`, `dob`

**Actions:**  
- Mask or remove fields  
- Flag event for PHI handling

---

## RET‑004 — Audit‑Critical Retention Override
**Category:** `retention_override`  
**Business Purpose:** Extend retention for events required for regulatory audit.

**Triggers:**  
- Event type in audit‑critical list  
- Tenant requires extended retention

**Applies To:**  
- **FHIR:** Claim, ExplanationOfBenefit  
- **X12:** 270/271, 276/277, 835, 837  
- **HL7:** ADT, ORU  
- **Generic JSON:** `audit_critical = true`

**Actions:**  
- Override retention to 7 years  
- `compliance_reason = "retention_override: audit_critical"`

---

## TENANT‑010 — Tenant Boundary Enforcement
**Category:** `tenant_restriction`  
**Business Purpose:** Prevent cross‑tenant data leakage.

**Triggers:**  
- Event contains `tenant_id`  
- Routing target does not match tenant

**Applies To:**  
- **All formats**

**Actions:**  
- Restrict routing  
- `compliance_flag = true`  
- `compliance_reason = "tenant_restriction"`

---

## CLASS‑020 — Classification Mismatch
**Category:** `classification_enforcement`  
**Business Purpose:** Detect when producers incorrectly classify PHI/PII as `internal` or `public`.

**Triggers:**  
- PHI/PII detected  
- Declared classification is not `phi` or `pii`

**Applies To:**  
- **FHIR:** Any PHI‑containing resource  
- **HL7:** PID, OBX, ORC  
- **X12:** Subscriber loops, clinical attachments  
- **Generic JSON:** Any PHI fields

**Actions:**  
- Elevate retention to 7 years  
- Restrict routing  
- `compliance_flag = true`  
- `compliance_reason = "classification_mismatch"`

---

# Concrete Format‑Specific Compliance Rules  
*(Derived from `seed_compliance_rules.sql`)*

The following rules are **structural, format‑specific validation rules** that ensure incoming events meet the minimum required structure for their respective standards (X12, HL7, FHIR, Generic JSON). These rules complement the business‑level rules above.

Each rule includes:
- **Format**  
- **Business meaning**  
- **What the rule checks**  
- **Meaning of `compliance_flag`**  
- **Meaning of `reason_code`**  
- **Canonical compliance outcome mapping**

---

## X12 Rules

### **x12_837_required_segments_present**
**Entity:** member  
**Format:** X12 837  
**Business Purpose:** Ensures the 837 claim contains all mandatory subscriber and claim segments.  
**Checks:**  
- NM1*IL (subscriber)  
- CLM (claim)  
- Required loops for member identity  

**compliance_flag = true** → All required segments present  
**compliance_flag = false** → Missing required segments  
**reason_code:**  
- `MISSING_NM1_IL` → Subscriber identity segment missing  

**Canonical Outcome:**  
- `compliance_flag` reflects structural validity  
- `compliance_reason` populated when missing  
- Event may be routed to an error stream depending on severity  

---

### **billing_provider_npi_valid**
**Entity:** provider  
**Format:** X12 837  
**Business Purpose:** Ensures the billing provider NPI is present and structurally valid.  
**Checks:**  
- NPI is 10 digits  
- NPI is present in the billing provider loop  

**compliance_flag = true** → Valid NPI  
**compliance_flag = false** → Invalid or missing NPI  

**Canonical Outcome:**  
- `compliance_flag = true` indicates structural compliance  
- Invalid NPI may restrict routing or trigger remediation  

---

### **x12_encounter_requires_clm**
**Entity:** encounter  
**Format:** X12 837  
**Business Purpose:** Ensures encounter‑level events include a CLM segment.  
**Checks:**  
- CLM segment exists  
- Encounter ID matches expected structure  

**compliance_flag = true** → Encounter is structurally valid  
**compliance_flag = false** → Missing CLM  

---

## HL7 Rules

### **hl7_pid_required**
**Entity:** encounter  
**Format:** HL7 ADT  
**Business Purpose:** Ensures ADT messages include a PID segment for patient identity.  
**Checks:**  
- PID segment present  
- Required PID fields populated  

**compliance_flag = true** → PID present  
**compliance_flag = false** → PID missing  

**Canonical Outcome:**  
- Missing PID triggers `compliance_flag = true` (violation)  
- Routing may be restricted  

---

### **hl7_patient_id_valid**
**Entity:** patient  
**Format:** HL7 ADT  
**Business Purpose:** Ensures patient identifiers in PID are valid.  
**Checks:**  
- PID‑3 populated  
- Identifier format valid  

**compliance_flag = true** → Patient ID valid  
**compliance_flag = false** → Invalid or missing ID  

---

## FHIR Rules

### **fhir_encounter_status_required**
**Entity:** encounter  
**Format:** FHIR Encounter  
**Business Purpose:** Ensures Encounter.status is present and valid.  
**Checks:**  
- `Encounter.status` exists  
- Value is one of: `planned`, `in-progress`, `finished`, etc.  

**compliance_flag = true** → Status valid  
**compliance_flag = false** → Missing or invalid status  

---

### **fhir_patient_reference_valid**
**Entity:** patient  
**Format:** FHIR Encounter  
**Business Purpose:** Ensures Encounter.subject references a valid Patient.  
**Checks:**  
- `Encounter.subject.reference` present  
- Reference resolves to a known Patient  

**compliance_flag = true** → Reference valid  
**compliance_flag = false** → Missing or invalid reference  

---

## Generic JSON Rules

### **generic_encounter_id_required**
**Entity:** encounter  
**Format:** JSON  
**Business Purpose:** Ensures generic encounter events include an encounter identifier.  
**Checks:**  
- `encounter_id` present  
- Non‑empty string  

**compliance_flag = true** → ID present  
**compliance_flag = false** → Missing ID  

---

### **generic_member_id_valid**
**Entity:** member  
**Format:** JSON  
**Business Purpose:** Ensures member identifiers meet expected format.  
**Checks:**  
- `member_id` present  
- Matches expected pattern (e.g., alphanumeric)  

**compliance_flag = true** → Valid member ID  
**compliance_flag = false** → Invalid or missing ID  

---

# Cross‑Format Applicability

## FHIR Events
FHIR resources are structured and typed, enabling precise PHI detection.

Examples:
- `Patient.name`, `Patient.address` → demographic masking  
- `Observation.code` → lab masking  
- `Claim` → retention override  

FHIR is the most expressive format; PHI rules fire frequently.

---

## Generic JSON Events
Generic JSON relies on naming conventions and schema metadata.

Examples:
- `dob`, `address`, `ssn` → demographic masking  
- `lab_code` → lab masking  
- `audit_critical = true` → retention override  

Generic JSON is the most flexible; rules rely heavily on schema metadata.

---

## X12 Events
X12 is positional; PHI appears in subscriber loops and attachments.

Examples:
- 270/271 eligibility → retention override  
- 837 claims → retention override  
- Subscriber loops → demographic masking  

X12 is rigid; rules rely on canonicalized fields extracted during parsing.

---

## HL7 v2 Events
HL7 v2 is segment‑based and PHI‑dense.

Examples:
- PID → demographic masking  
- OBX → lab masking  
- ORU/ADT → retention override  

HL7 is the most PHI‑dense format; PHI rules fire frequently.

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
```
