# HESP Event Lifecycle Model

**Platform:** Healthcare Event Stream Platform (HESP)  
**Audience:** Internal Aetna developers and platform engineers  
**Status:** Canonical reference — do not modify without Platform Architecture review  
**Last reviewed:** 2026-05

---

## Overview

This document defines the end-to-end lifecycle of an event on HESP — from construction by a producer, through platform ingestion, stream routing, consumption, and eventual archival or expiration. Understanding the lifecycle is essential for building producers and consumers that handle failures, replays, and ordering guarantees correctly.

---

## Lifecycle Stages

An event passes through six sequential stages. Each stage has a defined entry condition, platform behavior, and exit state.

```
CONSTRUCTED → SUBMITTED → INGESTED → ROUTED → CONSUMED → ARCHIVED
                  ↓
              REJECTED  (terminal, not retried by platform)
```

---

### Stage 1 — CONSTRUCTED

The event exists only in producer memory. No platform involvement.

**Producer obligations at this stage:**

- Generate a UUID v4 `event_id`.
- Set `produced_at` to the current UTC time.
- Apply field-level encryption to all PHI/PII payload fields per the PHI Handling Guide.
- Populate all required envelope fields.
- Do not set `ingested_at` — this field is platform-owned.

**Exit condition:** Producer calls the ingestion API.

---

### Stage 2 — SUBMITTED

The event has been received by the ingestion API but not yet written to a stream.

**Platform behavior at this stage:**

- Validates the envelope synchronously (see Validation and Rejection in the Event Model reference).
- Checks `source_system` against the Service Registry.
- Performs deduplication check on `event_id`. If a duplicate is detected, the API returns HTTP 200 (idempotent accept) without re-writing the event.
- Does not validate payload content.

**Exit conditions:**

- Envelope valid → transitions to INGESTED.
- Envelope invalid → transitions to REJECTED. HTTP 422 returned to producer.
- Duplicate `event_id` → HTTP 200 returned; event is not re-ingested.

**Producer guidance:** Producers should treat any non-200/422 response (e.g. 500, 503, timeout) as a signal to retry with exponential backoff. The `event_id` deduplication guarantee makes retries safe.

---

### Stage 3 — INGESTED

The event has been written durably to the HESP stream.

**Platform behavior at this stage:**

- Sets `ingested_at` timestamp on the envelope (immutable after this point).
- Assigns the event to a stream partition based on `tenant_id` and `event_type`.
- Appends a monotonically increasing stream offset within the partition.
- Guarantees durability: the event is replicated across availability zones before the HTTP 201 response is returned to the producer.

**Platform guarantees:**

- At-least-once delivery to the stream. Exactly-once delivery is not guaranteed at this stage; consumers must handle duplicates using `event_id`.
- Ordering is guaranteed within a partition (same `tenant_id` + `event_type`). Cross-partition ordering is not guaranteed.

**Exit condition:** Event is available for routing.

---

### Stage 4 — ROUTED

The platform has made the event available on the appropriate topic for registered consumers.

**Platform behavior at this stage:**

- Applies topic routing rules based on `event_type` and `data_classification`.
- Enforces stream ACLs: consumers without an explicit grant for the topic cannot read PHI/PII events.
- Applies content-based filter rules if configured by the consumer subscription (e.g. route only `claims.adjudication.completed` events where `tenant_id` matches a specific value).

**Consumer guidance:** Consumers that fall behind their stream offset are not dropped. HESP retains events on the stream for the duration of the `retention_policy` specified in the envelope, with a platform minimum of 90 days regardless of the declared policy.

---

### Stage 5 — CONSUMED

A registered consumer has read and acknowledged the event.

**Platform behavior at this stage:**

- Records the consumer group's committed offset.
- Does not delete the event from the stream upon acknowledgement. Events remain available for replay until the retention policy expires.

**Consumer obligations at this stage:**

- Acknowledge only after the event has been successfully processed and any downstream writes are durable. Do not acknowledge speculatively.
- Implement idempotent processing keyed on `event_id`. The platform may deliver the same event more than once in failure scenarios.
- Honor `data_classification`. PHI/PII events must not be written to systems without an approved data store designation.

**Replay:** Any consumer may reset its offset to any point within the retention window to replay events. Replays do not change the event's lifecycle state on the platform.

---

### Stage 6 — ARCHIVED

The event has exceeded its active stream retention window and has been moved to the HESP Cold Archive.

**Platform behavior at this stage:**

- Event is no longer available via the standard stream consumer API.
- Event is queryable via the HESP Archive Query API (higher latency, requires explicit access grant).
- Archive retention follows the `retention_policy` in the envelope. For `phi` and `pii` events, the minimum archive retention is 7 years in compliance with HIPAA §164.530(j).
- At the end of the archive retention period, the event is cryptographically erased (key destruction, not byte overwrite) to satisfy right-to-deletion obligations where applicable.

---

### Terminal State — REJECTED

An event that failed envelope validation during Stage 2. It was never written to a stream.

**Platform behavior:**

- The ingestion API returns HTTP 422 with a structured error body identifying the failing fields.
- The rejected event is written to the producer's Dead Letter endpoint (if configured in the Service Registry) for observability. The Dead Letter record contains the raw submitted payload and the rejection reason.
- HESP does not retry rejected events. Rejection is always due to a producer-side error.

**Producer guidance:** Monitor the Dead Letter endpoint. Persistent rejections indicate a schema mismatch, misconfigured envelope, or unregistered `source_system`. Do not build retry loops around 422 responses without first resolving the root cause.

---

## Ordering and Delivery Guarantees Summary

| Guarantee | Scope | Notes |
|---|---|---|
| Durable write | Per event | Replicated before HTTP 201 is returned |
| At-least-once delivery | Per partition | Consumers must be idempotent on `event_id` |
| Ordered delivery | Within partition | Partitioned by `tenant_id` + `event_type` |
| Cross-partition ordering | Not guaranteed | Use `correlation_id` to reconstruct transaction order |
| Exactly-once delivery | Not provided | Not available at platform level |
| Duplicate suppression | Ingestion only | `event_id` deduplication at Stage 2 only |

---

## Retention Policy Reference

| Policy value | Active stream retention | Archive retention | Applicable to |
|---|---|---|---|
| `7y` | 90 days | 7 years | Required for `phi`, `pii` |
| `3y` | 90 days | 3 years | Internal business records |
| `1y` | 90 days | 1 year | Operational events |
| `90d` | 90 days | None | Ephemeral / debug events |

All events, regardless of declared policy, are retained on the active stream for a platform minimum of 90 days to support consumer lag recovery.

---

## Failure Handling Responsibilities

| Scenario | Responsible party | Recommended action |
|---|---|---|
| Ingestion API 5xx / timeout | Producer | Retry with exponential backoff; `event_id` makes retries safe |
| Ingestion API 422 (rejection) | Producer | Inspect Dead Letter record; fix envelope before retrying |
| Consumer processing failure | Consumer | Do not acknowledge; allow re-delivery; implement dead letter consumer |
| Duplicate event received | Consumer | Deduplicate on `event_id` before processing |
| Consumer falls behind | Consumer | HESP retains events for 90-day minimum; catch up before window closes |
| PHI delivered to wrong system | Consumer | Incident response per the Aetna Data Breach Policy; notify HESP Access team |
