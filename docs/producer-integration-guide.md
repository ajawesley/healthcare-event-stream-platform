# AcmeCo Event Stream Platform — Producer Integration Guide

**Platform:** AcmeCo Event Stream Platform (ESP)  
**Audience:** Engineering teams producing events into ESP  
**Status:** Canonical reference — do not modify without Platform Architecture review  
**Last reviewed:** 2026‑05

---

## Overview

This guide provides engineering teams with the requirements, responsibilities, and best practices for producing events into the AcmeCo Event Stream Platform (ESP). It covers:

- producer responsibilities  
- envelope construction  
- payload schema registration  
- PHI/PII handling  
- ingestion API usage  
- idempotency and retries  
- validation and rejection rules  
- local and integration testing  

This guide is mandatory for all teams onboarding as producers.

---

## Producer Responsibilities

Producers are responsible for constructing valid events and submitting them to the ESP ingestion API. Producers must:

- generate UUID v4 `event_id`  
- set `produced_at` in UTC  
- classify data correctly (`phi`, `pii`, `internal`, `public`)  
- encrypt PHI/PII fields before ingestion  
- populate all required envelope fields  
- validate payload schema against the ESP Schema Registry  
- register their `source_system`  
- implement safe retry logic using `event_id`  
- avoid sending PHI in logs or metadata  

Producers **must not**:

- set `ingested_at`  
- send unencrypted PHI  
- misclassify PHI as `internal` or `public`  
- bypass the ingestion API  
- embed PHI in envelope or metadata fields  

---

## Event Construction

Every event must contain:

1. **Envelope** — platform‑owned metadata  
2. **Payload** — domain‑owned business data  
3. **Optional canonical structures** — added by ESP  
4. **Compliance metadata** — added by ESP  

Producers are responsible only for the envelope and payload.

---

## Envelope Requirements

All required envelope fields must be present and valid.

| Field | Required | Description |
|---|---|---|
| `event_id` | Yes | UUID v4, used for idempotency |
| `event_type` | Yes | `<domain>.<entity>.<verb>` |
| `event_version` | Yes | Semantic version of payload schema |
| `produced_at` | Yes | ISO 8601 UTC timestamp |
| `source_system` | Yes | Registered system identifier |
| `correlation_id` | Yes | Groups related events |
| `tenant_id` | Yes | AcmeCo business unit or segment |
| `data_classification` | Yes | `phi`, `pii`, `internal`, `public` |
| `retention_policy` | Optional | Defaults based on classification |
| `schema_registry_url` | Optional | Link to schema definition |

### Naming Rules for `event_type`

Must follow:

```
<domain>.<entity>.<verb>
```

Examples:

- `claims.adjudication.completed`
- `member.enrollment.updated`
- `provider.directory.failed`

Invalid:

- `MemberEnrollmentCompleted`
- `member-enrollment-completed`
- `enrollment.complete`

---

## Payload Requirements

The payload:

- must be a JSON object  
- must conform to the registered schema  
- must not contain a top‑level `envelope` field  
- must encrypt PHI/PII fields before ingestion  
- must not contain unencrypted identifiers  
- must not contain PHI in metadata fields  

Payload schemas are versioned and validated through the ESP Schema Registry.

---

## Schema Registry Integration

Before producing events, teams must:

1. Define or update the payload schema  
2. Register the schema in the ESP Schema Registry  
3. Validate schema compatibility  
4. Link schema to `event_type` + `event_version`  
5. Pass CI schema validation checks  

Schema registration is a deployment gate.

---

## PHI/PII Handling Requirements

Producers must:

- encrypt PHI/PII fields using approved encryption libraries  
- avoid placing PHI in envelope fields  
- avoid placing PHI in metadata fields  
- avoid logging PHI at any time  

ESP does **not** encrypt payload fields on behalf of producers.

---

## Ingestion API

Producers submit events via the ESP ingestion API using mTLS.

### Request Format

```
POST /v1/events
Content-Type: application/json
```

Body:

```json
{
  "envelope": { ... },
  "payload": { ... }
}
```

### Response Codes

| Code | Meaning |
|---|---|
| `202 Accepted` | Event accepted for processing |
| `400 Bad Request` | Invalid JSON |
| `401 Unauthorized` | mTLS failure |
| `422 Unprocessable Entity` | Envelope validation failed |
| `500 Internal Server Error` | Platform error (retry safe) |

---

## Idempotency & Retry Logic

Producers must implement safe retry logic using `event_id`.

### Rules:

- Reuse the same `event_id` for all retries  
- Do not generate a new ID on retry  
- Retries must be exponential backoff  
- Retries must stop after a bounded number of attempts  

ESP guarantees:

- duplicate events with the same `event_id` are deduplicated  
- ingestion is at‑least‑once  

---

## Validation & Rejection Rules

Events are rejected if:

- required envelope fields are missing  
- `event_id` is not UUID v4  
- `event_type` violates naming rules  
- `data_classification` is invalid  
- `ingested_at` is present  
- `produced_at` is more than 5 minutes in the future  
- `source_system` is not registered  
- payload schema validation fails  

Rejected events are **not** written to any stream.

---

## Local Testing

Producers must validate events locally using:

- envelope validator  
- schema validator  
- compliance mock  
- ingestion mock  
- replay simulator (optional)  

Local testing ensures producers catch issues before integration.

---

## Integration Testing

Integration tests run in the ESP Shared Services environment:

- ingestion → canonicalization → compliance → routing  
- schema registry integration  
- PHI masking validation  
- retention policy validation  

Teams must pass integration tests before production approval.

---

## Producer Readiness Checklist

| Requirement | Status |
|---|---|
| Envelope fields implemented | ☐ |
| Payload schema registered | ☐ |
| PHI encrypted | ☐ |
| Source system registered | ☐ |
| Retry logic implemented | ☐ |
| Local tests passing | ☐ |
| Integration tests passing | ☐ |
| No PHI in logs | ☐ |
| No PHI in metadata | ☐ |

---

## Summary

This guide defines the requirements for producing events into the AcmeCo Event Stream Platform. By following these standards, producers ensure:

- safe ingestion  
- correct classification  
- schema compliance  
- PHI/PII protection  
- reliable retries  
- consistent event quality  

Producers play a critical role in maintaining the integrity and compliance posture of the platform.
