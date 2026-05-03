# HESP Ingest Service — MVP Architecture

**Scope:** MVP ingestion service only  
**Deployment target:** AWS (ECS Fargate preferred; Lambda noted where applicable)  
**Status:** Draft — pending Platform Architecture review  
**Last reviewed:** 2026-05

---

## Purpose

This document describes the architecture of the HESP MVP ingest service. The scope is deliberately narrow: accept an event over HTTP, validate the envelope, write the raw payload to S3, and return a correlation ID to the caller. Nothing else.

Routing, lifecycle assignment, schema enforcement, canonical normalization, and HIPAA guardrails are explicitly out of scope for this phase and are not referenced here.

---

## System Boundary

The MVP consists of four components. Everything outside this boundary — downstream consumers, normalization pipelines, and the schema registry — is a future concern.

```
Client
  │
  │  POST /events/ingest
  ▼
API Gateway / ALB          ← public entrypoint, TLS termination
  │
  ▼
Ingest Service (Go)        ← envelope validation, S3 write, response
  │
  ├──▶ S3 Bucket           ← durable raw event storage
  │
  └──▶ CloudWatch          ← structured logs, basic metrics
```

---

## Components

### Ingest Service (Go)

The single deployable unit. Runs as a container on ECS Fargate. Stateless — no local storage, no in-memory queue.

**Endpoints:**

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/healthz` | Liveness probe. Returns `200 OK` with no body. Used by ALB target group health checks. |
| `POST` | `/events/ingest` | Accepts a single JSON event, validates the envelope, writes to S3, returns `202 Accepted`. |

**Responsibilities:**

- Validate that required envelope fields are present and well-formed (see Envelope Validation below).
- Generate an `ingested_at` timestamp and attach it to the S3 object metadata.
- Write the raw request body to S3 without modification.
- Return a `202 Accepted` response containing the `event_id` and `ingested_at` values.
- Emit a structured log line and a CloudWatch metric on every ingest attempt, regardless of outcome.

**Explicitly not responsible for:**

- Payload parsing or validation.
- Deduplication.
- Routing or topic assignment.
- Any downstream processing trigger.

**Runtime configuration** (environment variables):

| Variable | Required | Description |
|---|---|---|
| `HESP_S3_BUCKET` | Yes | Target S3 bucket name. |
| `HESP_S3_PREFIX` | No | Key prefix for stored objects. Defaults to `raw/`. |
| `HESP_ENV` | Yes | Deployment environment: `dev`, `staging`, `prod`. Included in log output. |
| `HESP_LOG_LEVEL` | No | Log verbosity: `debug`, `info`, `warn`, `error`. Defaults to `info`. |

---

### API Gateway / ALB

Public entrypoint. Terminates TLS and forwards requests to the ingest service.

**MVP preference:** Application Load Balancer (ALB) with an ECS Fargate target group. ALB adds less latency than API Gateway for a single-route service and has a simpler operational model at this stage.

**API Gateway** is a valid alternative if rate limiting or AWS WAF integration is required before launch. The ingest service is agnostic to which sits in front of it.

Both options terminate TLS at the boundary. The ingest service communicates over plain HTTP within the VPC.

---

### S3 Bucket

Durable storage for raw event payloads. No processing occurs in S3 at this phase — it is a landing zone only.

**Object key format:**

```
{prefix}/{YYYY}/{MM}/{DD}/{event_id}.json
```

Example: `raw/2026/05/01/a3f1c847-2d09-4e6b-91f4-bc3301d78e12.json`

Date-partitioned keys make it straightforward to list, audit, or reprocess events by day without scanning the full bucket.

**Object metadata** (set by the ingest service, not the producer):

| Key | Value |
|---|---|
| `x-hesp-event-id` | Value of `envelope.event_id` from the request |
| `x-hesp-event-type` | Value of `envelope.event_type` from the request |
| `x-hesp-ingested-at` | ISO 8601 UTC timestamp set by the ingest service |
| `x-hesp-source-system` | Value of `envelope.source_system` from the request |

**Bucket configuration:**

- Server-side encryption enabled (SSE-S3 for MVP; SSE-KMS upgrade path documented but not required at this phase).
- Versioning enabled.
- Public access blocked.
- Lifecycle rule: transition objects to S3 Intelligent-Tiering after 30 days.

---

### CloudWatch

Operational visibility for the MVP. No custom dashboards required at this phase — structured logs and a small set of metrics are sufficient.

**Log format:** JSON, one line per event, emitted to a CloudWatch Log Group named `/hesp/ingest/{env}`.

Every log line includes: `event_id`, `event_type`, `source_system`, `ingested_at`, `outcome` (`accepted` or `rejected`), `http_status`, and `duration_ms`.

**Metrics** (CloudWatch custom namespace `HESP/Ingest`):

| Metric | Unit | Description |
|---|---|---|
| `IngestAttempts` | Count | Total ingest requests received. |
| `IngestAccepted` | Count | Requests that passed validation and were written to S3. |
| `IngestRejected` | Count | Requests that failed envelope validation. |
| `S3WriteLatency` | Milliseconds | Time to complete the S3 `PutObject` call. |
| `HandlerLatency` | Milliseconds | Total time from request receipt to response sent. |

Alarms on `IngestRejected` (threshold: >10 in 5 minutes) and `S3WriteLatency` (threshold: p99 > 2000 ms) should be configured before production traffic is enabled.

---

## Request and Response Contract

### Request

```
POST /events/ingest
Content-Type: application/json
```

Body: a single JSON object containing at minimum an `envelope` field. The payload field is accepted but not inspected.

### Successful Response — 202 Accepted

```json
{
  "event_id": "a3f1c847-2d09-4e6b-91f4-bc3301d78e12",
  "ingested_at": "2026-05-01T14:23:00Z"
}
```

### Rejection Response — 422 Unprocessable Entity

```json
{
  "error": "envelope_validation_failed",
  "fields": ["envelope.event_id", "envelope.produced_at"]
}
```

### Error Response — 500 Internal Server Error

```json
{
  "error": "internal_error",
  "message": "failed to write to storage"
}
```

The `event_id` is never included in a 500 response — the write may not have completed.

---

## Envelope Validation

The ingest service checks only that required envelope fields are present and have the correct types. No semantic validation (e.g. whether `source_system` is registered, whether `event_type` follows naming conventions) is performed in the MVP.

**Required fields checked at MVP:**

| Field | Type check |
|---|---|
| `envelope.event_id` | Non-empty string |
| `envelope.event_type` | Non-empty string |
| `envelope.produced_at` | Non-empty string, parseable as ISO 8601 |
| `envelope.source_system` | Non-empty string |

Any field that is missing, null, or the wrong type causes the entire request to be rejected with a 422. The response body lists all failing fields so the producer can correct them in a single round trip.

---

## Deployment

**ECS Fargate task configuration (baseline):**

| Parameter | MVP value |
|---|---|
| CPU | 512 units (0.5 vCPU) |
| Memory | 1024 MB |
| Desired count | 2 (one per AZ) |
| Auto-scaling | Not enabled for MVP |

**IAM permissions required by the task role:**

- `s3:PutObject` on the target bucket and prefix.
- `cloudwatch:PutMetricData` on the `HESP/Ingest` namespace.
- `logs:CreateLogStream`, `logs:PutLogEvents` on the `/hesp/ingest/*` log group.

No other AWS permissions are required or should be granted.

---

## What Comes Next

The MVP establishes the landing zone. The following capabilities are planned for subsequent phases but are not referenced in this document:

- Envelope semantic validation (service registry lookup, `event_type` naming rules).
- Deduplication on `event_id`.
- Stream routing (Kinesis or Kafka topic assignment).
- Event lifecycle state tracking.
- Schema registry integration.
- HIPAA technical safeguards (field-level encryption validation, PHI stream ACLs, audit trail).

When any of these are added, this document will be superseded by a revised architecture document scoped to that phase.

