# Canonical Event Model

The canonical event model provides a consistent, HIPAA‑aligned schema for all healthcare events ingested by the Healthcare Event Stream Platform (HESP). Regardless of source format (HL7, FHIR, X12, EHR webhook, proprietary), all events are normalized into this structure.

## 1. Event Schema

```jsonc
{
  "event_type": "string",          // e.g., encounter.updated, claim.submitted
  "patient_id": "string",          // hashed or tokenized identifier
  "provider_id": "string",         // hashed or tokenized identifier
  "encounter_id": "string|null",   // correlation identifier
  "lifecycle_state": "string",     // platform-assigned lifecycle state
  "timestamp": "ISO-8601 string",  // event occurrence time
  "raw_payload_ref": "string",     // S3 URI to encrypted raw payload
  "metadata": {
    "source_system": "string",
    "sequence": "number",
    "ingested_at": "ISO-8601 string",
    "phi_classification": "PHI|ePHI|non-PHI"
  }
}
```

## 2. Design Principles

- **Minimum necessary**: Only essential identifiers are retained; all PHI is hashed or redacted.
- **Source-agnostic**: HL7, FHIR, X12, and custom formats all map cleanly.
- **Lifecycle-aware**: Every event carries a lifecycle state.
- **Replay-safe**: Sequence numbers ensure deterministic reprocessing.
- **Audit-ready**: Metadata supports lineage, compliance, and traceability.

## 3. Identifier Handling

- `patient_id` and `provider_id` are hashed using a platform-managed salt.
- Raw identifiers never appear in canonical events.
- All PHI is stored only in encrypted raw payloads.

## 4. Extensibility

The model supports future expansion:
- additional lifecycle states
- domain-specific metadata
- new event types
- new ingestion sources

This schema is the foundation for all downstream processing, storage, and analytics.

