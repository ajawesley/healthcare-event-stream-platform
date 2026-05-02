# HESP HIPAA Alignment Summary

**Platform:** Healthcare Event Stream Platform (HESP)  
**Audience:** Internal Aetna developers, platform engineers, and compliance reviewers  
**Status:** Canonical reference — do not modify without Platform Architecture and Compliance review  
**Last reviewed:** 2026-05

---

## Purpose

This document maps HESP platform controls to relevant HIPAA Security Rule and Privacy Rule requirements. It is intended to help development teams understand which obligations the platform satisfies on their behalf, and which obligations remain with the producer or consumer service.

This document is a developer-facing alignment summary, not a formal compliance attestation. The authoritative compliance record is maintained by the Aetna Privacy and Security Office.

---

## Scope

This summary covers events classified as `phi` or `pii` in the HESP envelope `data_classification` field. Events classified as `internal` or `public` are not subject to HIPAA controls, though platform security controls still apply.

---

## Control Mapping

### Technical Safeguards — 45 CFR §164.312

**Access Control (§164.312(a))**

The platform satisfies this requirement through stream-level ACLs enforced at the routing layer. No consumer can read a PHI or PII stream without an explicit access grant issued by the HESP Access team. Access grants are tied to service identity, not individual credentials, and are reviewed quarterly.

Developer obligation: Request stream access through the HESP Access team before deploying a PHI consumer. Do not share service credentials between systems.

**Audit Controls (§164.312(b))**

Every read and write operation against a PHI or PII event is recorded in the HESP Audit Trail. Audit records are immutable, tamper-evident, and retained for 7 years. The audit trail captures the `event_id`, `source_system` or consuming service identity, operation type, and timestamp.

Developer obligation: None — the platform records audit events automatically. Do not attempt to suppress or filter audit trail output.

**Integrity Controls (§164.312(c))**

Events are written with a cryptographic checksum at ingestion. The platform verifies this checksum at each stage of the lifecycle. Events that fail integrity verification are quarantined and flagged for investigation. The stream offset is monotonically increasing and append-only; events cannot be modified after ingestion.

Developer obligation: Do not attempt to mutate an event after ingestion. To correct an error, produce a new corrective event with a new `event_id` and a `causation_id` referencing the original.

**Transmission Security (§164.312(e))**

All ingestion API endpoints require mutual TLS (mTLS). Plain HTTP connections are rejected at the load balancer. In-transit data is encrypted using TLS 1.2 minimum; TLS 1.3 is recommended and the default for all new service integrations.

Developer obligation: Configure your HTTP client to present a valid client certificate issued by the Aetna Internal CA. Certificates are provisioned through the HESP Onboarding process.

**Encryption at Rest**

All events stored in the HESP stream and archive are encrypted at rest using AES-256 with Aetna-managed keys stored in the enterprise key management service. Key rotation occurs annually or on-demand following a security event.

Developer obligation: PHI field values within the payload must be field-level encrypted by the producer before submission to the ingestion API, using the scheme defined in the PHI Handling Guide. The platform encrypts the full event envelope and payload at rest, but field-level encryption provides an additional layer of protection in the event of a partial platform compromise.

---

### Administrative Safeguards — 45 CFR §164.308

**Access Management (§164.308(a)(4))**

Stream access grants are role-based and follow the principle of least privilege. Grants are scoped to specific `event_type` namespaces and `tenant_id` values. Broad wildcard grants are not issued without explicit approval from the HESP Access team and the Privacy Office.

Developer obligation: Request only the stream access your service needs. Access requests must include the specific `event_type` patterns and `tenant_id` values required.

**Contingency Plan (§164.308(a)(7))**

HESP is deployed across multiple availability zones with automatic failover. Event data is replicated synchronously before a write is acknowledged. The Cold Archive provides a secondary copy of all PHI events for disaster recovery purposes. Recovery Time Objective (RTO) and Recovery Point Objective (RPO) targets are documented in the HESP Service Level Agreement.

Developer obligation: Design producers and consumers to tolerate transient ingestion API unavailability using retry logic with exponential backoff. Do not assume synchronous end-to-end delivery.

**Evaluation (§164.308(a)(8))**

The HESP platform undergoes annual security review, penetration testing, and HIPAA compliance assessment conducted by the Aetna Security team in coordination with the Privacy Office. Results are shared with platform consumers through the internal security advisory process.

---

### Privacy Rule — 45 CFR §164.502 and §164.514

**Minimum Necessary Standard (§164.502(b))**

The `data_classification` and stream ACL model enforces the minimum necessary principle at the platform level. A service processing claims adjudication events cannot read member enrollment events unless it holds a separate access grant for that stream.

Developer obligation: Do not request access to PHI streams beyond what your service's documented purpose requires. Over-broad access grants will be rejected during the access review process.

**De-identification (§164.514(b))**

HESP does not perform automatic de-identification. If a use case requires publishing de-identified data derived from PHI events, the producing service must perform de-identification before constructing the event, and must classify the resulting event as `internal` rather than `phi`.

Developer obligation: Do not submit partially de-identified events as `phi`. Consult the Privacy Office to confirm that your de-identification method satisfies Safe Harbor or Expert Determination standards before reclassifying.

**Right to Deletion**

At the end of the archive retention period, PHI events are cryptographically erased via key destruction rather than byte-level overwrite. This approach satisfies deletion obligations while maintaining the integrity of append-only audit logs.

Developer obligation: If a member exercises a deletion right and your service holds derived data sourced from HESP events, you are responsible for handling deletion of that derived data. The HESP platform handles only the raw event archive.

---

## Platform Obligations vs. Developer Obligations — Quick Reference

| Control area | Platform handles | Developer handles |
|---|---|---|
| Encryption in transit | mTLS enforcement at ingestion | Client certificate provisioning |
| Encryption at rest | AES-256 full-event encryption | Field-level PHI encryption in payload |
| Access control | Stream ACLs, grant enforcement | Access requests scoped to need |
| Audit logging | Immutable audit trail, 7-year retention | None |
| Event integrity | Checksum verification, append-only stream | Corrective events instead of mutations |
| Retention and archival | Automated per `retention_policy` | Setting the correct `retention_policy` in envelope |
| De-identification | Not provided | Producer-side, before event construction |
| Duplicate delivery handling | `event_id` dedup at ingestion | Consumer-side idempotency on `event_id` |
| Breach notification | Platform-level incident response | Service-level incident reporting to HESP team |

---

## Contacts

| Need | Contact |
|---|---|
| Stream access grants | HESP Access team — `#hesp-access` in Slack |
| PHI handling guidance | Aetna Privacy Office |
| Security incidents | Aetna Security Operations Center |
| Platform onboarding | HESP Platform Engineering — `#hesp-platform` in Slack |
| Compliance questions | `hesp-compliance@aetna.internal` |
