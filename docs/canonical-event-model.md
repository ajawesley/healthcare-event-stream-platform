# ESP Canonical Event Model

**Platform:** AcmeCo Event Stream Platform (ESP)  
**Audience:** Internal AcmeCo developers and platform engineers  
**Status:** Canonical reference — do not modify without Platform Architecture review  
**Last reviewed:** 2026‑05

---

## Overview

Every record flowing through the AcmeCo Event Stream Platform must conform to the canonical event schema defined in this document. The schema enforces a strict separation between platform‑level envelope fields, business‑domain payload fields, normalized canonical structures, and platform‑applied compliance metadata.

This separation allows the platform to route, audit, replay, and apply compliance rules without parsing domain‑specific content, while allowing domain teams to evolve their payloads independently.

---

## Structural Contract

An ESP event is a JSON object composed of two required top‑level sections — the **envelope** and the **payload** — plus optional canonical structures and platform‑applied compliance metadata.

```
{
  "envelope":  { ... },   // platform-owned — never modified by producers
  "payload":   { ... },   // domain-owned — schema varies by event_type
  "patient":   { ... },   // optional normalized structure
  "encounter": { ... },   // optional normalized structure
  "observation": { ... }, // optional normalized structure
  "metadata":  { ... },   // optional platform metadata
  "raw_value": { ... },   // optional lineage/debugging
  "compliance_applied": true,
  "compliance_flag": false,
  "compliance_reason": "...",
  "compliance_rule_type": "...",
  "compliance_rule_id": "...",
  "compliance_timestamp": "..."
}
```

Producers are responsible only for constructing a valid envelope and payload. Canonical structures and compliance metadata are added by the platform’s transformation and compliance layers.

---

## Envelope Fields

All envelope fields are required unless marked optional.

| Field | Type | Description |
|---|---|---|
| `event_id` | string (UUID v4) | Globally unique identifier for this event. Used for idempotency and deduplication. |
| `event_type` | string | Dot‑namespaced type identifier. Format: `<domain>.<entity>.<verb>`. |
| `event_version` | string (semver) | Schema version of the payload. |
| `produced_at` | string (ISO 8601 UTC) | Timestamp when the producer created the event. Must be UTC. |
| `ingested_at` | string (ISO 8601 UTC) | Set by the ingestion API. Producers must not set this. |
| `source_system` | string | Canonical identifier of the producing system. Must be registered. |
| `correlation_id` | string | Groups related events across a business transaction. |
| `causation_id` | string (optional) | The `event_id` of the event that directly caused this one. |
| `tenant_id` | string | AcmeCo line‑of‑business or market segment identifier. |
| `data_classification` | string (enum) | One of: `phi`, `pii`, `internal`, `public`. |
| `retention_policy` | string (enum) | One of: `7y`, `3y`, `1y`, `90d`. Defaults to `7y` for PHI/PII. |
| `schema_registry_url` | string (URI, optional) | URL to the payload schema in the ESP Schema Registry. |

### event_type Naming Convention

Event types must follow the pattern `<domain>.<entity>.<verb>` using lowercase snake_case segments separated by dots.

- **domain** — bounded context (e.g., `claims`, `member`, `provider`, `pharmacy`)  
- **entity** — primary domain object (e.g., `adjudication`, `enrollment`, `authorization`)  
- **verb** — past‑tense action (e.g., `created`, `updated`, `completed`, `failed`)  

Valid: `member.enrollment.completed`  
Invalid: `MemberEnrollmentCompleted`, `member-enrollment-completed`, `enrollment.complete`

---

## Payload Fields

The payload schema is defined and versioned by the owning domain team and registered in the ESP Schema Registry. The platform makes no assumptions about payload structure beyond the following constraints:

- The payload must be a JSON object.  
- The payload must not include a top‑level field named `envelope`.  
- PHI/PII fields must be field‑level encrypted before reaching the ingestion API.  
- The payload must conform to the schema registered for its `event_type` and `event_version`.  

The platform does not perform field‑level encryption on behalf of producers.

---

## Canonical Structures

The transformation layer produces normalized structures that downstream consumers can rely on, regardless of source format.

### CanonicalPatient

| Field | Type |
|---|---|
| `id` | string |
| `first_name` | string |
| `last_name` | string |

### CanonicalEncounter

| Field | Type |
|---|---|
| `id` | string |
| `type` | string |

### CanonicalObservation

| Field | Type |
|---|---|
| `code` | string |
| `value` | any |

These structures are optional and appear only when the source event contains the corresponding domain information.

---

## Compliance Metadata

The Compliance Engine enriches every canonical event with platform‑applied compliance metadata. These fields are always present after transformation.

| Field | Type | Description |
|---|---|---|
| `compliance_applied` | bool | Whether compliance rules were evaluated. |
| `compliance_flag` | bool | Whether any rule triggered. |
| `compliance_reason` | string | Human‑readable explanation of the rule outcome. |
| `compliance_rule_type` | string | Category of rule (e.g., `phi_masking`, `retention_override`). |
| `compliance_rule_id` | string | Unique identifier of the rule that fired. |
| `compliance_timestamp` | string (ISO 8601 UTC) | When compliance evaluation occurred. |

These fields correspond directly to the compliance metadata in the `CanonicalEvent` Go struct.

---

## Minimal Valid Event (Example)

```json
{
  "envelope": {
    "event_id": "a3f1c847-2d09-4e6b-91f4-bc3301d78e12",
    "event_type": "claims.adjudication.completed",
    "event_version": "2.1.0",
    "produced_at": "2026-05-01T14:23:00Z",
    "source_system": "claims-adjudication-service",
    "correlation_id": "txn-8821-acmeco-central",
    "tenant_id": "acmeco-commercial-midwest",
    "data_classification": "phi",
    "retention_policy": "7y"
  },
  "payload": {
    "claim_id": "CLM-20260501-99821",
    "member_id_encrypted": "<encrypted>",
    "adjudication_result": "approved",
    "allowed_amount_cents": 145000
  },
  "patient": {
    "id": "mem-99821"
  },
  "metadata": {
    "source_format": "hl7v2"
  },
  "compliance_applied": true,
  "compliance_flag": false,
  "compliance_reason": "",
  "compliance_rule_type": "",
  "compliance_rule_id": "",
  "compliance_timestamp": "2026-05-01T14:23:01Z"
}
```

---

## Schema Registry

All payload schemas must be registered in the ESP Schema Registry before a producer goes to production. The registry enforces:

- semantic versioning  
- backward‑compatibility checks  
- breaking change detection  
- linkage to the `event_type` + `event_version` composite key  

Schema registration is a deployment gate. CI pipelines must pass schema validation before publishing events.

---

## Validation and Rejection

The ingestion API performs envelope validation synchronously at receipt. Payload validation is the responsibility of the consuming service.

An event will be rejected (HTTP 422) if:

- any required envelope field is missing or null  
- `event_type` violates naming rules  
- `data_classification` or `retention_policy` is invalid  
- `ingested_at` is present in producer input  
- `event_id` is not a UUID v4  
- `produced_at` is more than 5 minutes in the future  
- `source_system` is not registered  

Rejected events are not written to any stream.

---

## PHI and PII Handling

Events classified as `phi` or `pii` are subject to:

- **Encryption in transit:** mTLS required  
- **Encryption at rest:** AES‑256 with AcmeCo‑managed keys  
- **Access control:** Stream ACLs enforced by the ESP Access service  
- **Audit logging:** All access logged and retained for 7 years  

Producers must not place PHI in fields classified as `internal` or `public`.
