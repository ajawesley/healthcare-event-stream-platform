# HESP Canonical Event Model

**Platform:** Healthcare Event Stream Platform (HESP)  
**Audience:** Internal Aetna developers and platform engineers  
**Status:** Canonical reference — do not modify without Platform Architecture review  
**Last reviewed:** 2026-05

---

## Overview

Every record flowing through HESP must conform to the canonical event schema defined in this document. The schema enforces a strict separation between platform-level envelope fields and business-domain payload fields. This separation allows the platform to route, audit, and replay events without parsing domain-specific content, and allows domain teams to evolve their payloads independently.

---

## Structural Contract

A HESP event is a JSON object composed of two top-level sections: the **envelope** and the **payload**.

```
{
  "envelope": { ... },   // platform-owned — never modified by producers
  "payload":  { ... }    // domain-owned — schema varies by event_type
}
```

Producers are responsible for constructing a valid envelope. The ingestion API will reject events with a missing or malformed envelope. The payload is opaque to the platform routing layer and is validated only by the consuming service.

---

## Envelope Fields

All envelope fields are required unless marked optional.

| Field | Type | Description |
|---|---|---|
| `event_id` | string (UUID v4) | Globally unique identifier for this event. Producers must generate this. The platform uses it for deduplication and idempotency. |
| `event_type` | string | Dot-namespaced type identifier. Format: `<domain>.<entity>.<verb>`. Example: `claims.adjudication.completed`. |
| `event_version` | string (semver) | Schema version of the payload. Example: `1.2.0`. The platform preserves this for consumers performing version negotiation. |
| `produced_at` | string (ISO 8601 UTC) | Timestamp at which the producer created the event. Must be UTC. Example: `2026-05-01T14:23:00Z`. |
| `ingested_at` | string (ISO 8601 UTC) | Set by the ingestion API at time of receipt. Producers must omit this field; the platform will reject events that include it. |
| `source_system` | string | Canonical system identifier of the producer. Registered in the HESP Service Registry. Example: `claims-adjudication-service`. |
| `correlation_id` | string | Identifier used to group related events across a business transaction. Producers should propagate this from upstream context (e.g. request trace ID). |
| `causation_id` | string (optional) | The `event_id` of the event that directly caused this one. Omit for root events. Used to reconstruct causal chains in audit logs. |
| `tenant_id` | string | Aetna line-of-business or market segment identifier. Required for multi-tenant stream partitioning. |
| `data_classification` | string (enum) | Sensitivity classification. Must be one of: `phi`, `pii`, `internal`, `public`. See the Data Classification Policy for definitions. |
| `retention_policy` | string (enum) | Governs how long the event is retained in the stream and archive. Must be one of: `7y` (HIPAA minimum), `3y`, `1y`, `90d`. Defaults to `7y` for `phi` and `pii` events. |
| `schema_registry_url` | string (URI, optional) | Fully qualified URL to the payload schema in the HESP Schema Registry. Recommended for `event_version` >= `1.0.0`. |

### event_type Naming Convention

Event types must follow the pattern `<domain>.<entity>.<verb>` using lowercase snake_case segments separated by dots.

- **domain** — the bounded context that owns the event. Example: `claims`, `member`, `provider`, `pharmacy`.
- **entity** — the primary domain object. Example: `adjudication`, `enrollment`, `authorization`.
- **verb** — past-tense action. Example: `created`, `updated`, `completed`, `failed`, `expired`.

Valid: `member.enrollment.completed`  
Invalid: `MemberEnrollmentCompleted`, `member-enrollment-completed`, `enrollment.complete`

---

## Payload Fields

The payload schema is defined and versioned by the owning domain team and registered in the HESP Schema Registry. The platform makes no assumptions about payload structure beyond the following constraints.

- The payload must be a JSON object. Scalar and array payloads are not accepted.
- The payload must not include a field named `envelope`. Namespace collisions at the top level will cause ingestion rejection.
- PHI and PII fields within the payload must be field-level encrypted before the event reaches the ingestion API. The platform does not perform field-level encryption on behalf of producers. See the PHI Handling Guide for the approved encryption scheme.

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
    "correlation_id": "txn-8821-aetna-central",
    "tenant_id": "aetna-commercial-midwest",
    "data_classification": "phi",
    "retention_policy": "7y"
  },
  "payload": {
    "claim_id": "CLM-20260501-99821",
    "member_id_encrypted": "<encrypted>",
    "adjudication_result": "approved",
    "allowed_amount_cents": 145000
  }
}
```

---

## Schema Registry

All payload schemas must be registered in the HESP Schema Registry before a producer goes to production. The registry enforces:

- Semantic versioning with backward-compatibility checks between minor versions.
- Breaking change detection on major version bumps (field removal, type changes, required field additions).
- Schema linkage to the `event_type` + `event_version` composite key in the envelope.

Schema registration is a deployment gate. CI pipelines that publish to HESP must pass schema validation as a required check.

---

## Validation and Rejection

The ingestion API performs envelope validation synchronously at the point of receipt. Payload validation is the responsibility of the consuming service.

An event will be rejected (HTTP 422) if any of the following are true:

- Any required envelope field is missing or null.
- `event_type` does not match the `<domain>.<entity>.<verb>` pattern.
- `data_classification` or `retention_policy` contains an unregistered value.
- `ingested_at` is present in the producer-submitted payload.
- `event_id` is not a valid UUID v4.
- `produced_at` is more than 5 minutes in the future (clock skew guard).
- `source_system` is not found in the HESP Service Registry.

Rejected events are not written to any stream. The producer is responsible for handling and retrying rejected events.

---

## PHI and PII Handling

Events classified as `phi` or `pii` are subject to the following platform-level controls in addition to producer obligations.

- **Encryption in transit:** All ingestion API endpoints require mTLS. HTTP is not accepted.
- **Encryption at rest:** The platform encrypts all stored events using AES-256 with Aetna-managed keys. Producers do not need to encrypt the envelope or the full payload; they are responsible only for field-level encryption of specific PHI/PII fields as defined in the PHI Handling Guide.
- **Access control:** PHI streams are accessible only to services with an explicit stream ACL grant, issued by the HESP Access team.
- **Audit logging:** All read and write access to PHI events is logged to the HESP Audit Trail, which is immutable and retained for 7 years.

Producers must not attempt to circumvent classification by placing PHI in fields classified as `internal` or `public`.
