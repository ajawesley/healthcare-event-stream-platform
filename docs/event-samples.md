```md
# HESP — End‑to‑End Event Samples (X12, HL7, FHIR, JSON)

This document provides complete examples for all supported formats:

- X12 (837)
- HL7 (ADT)
- FHIR (Encounter)
- Generic JSON (REST)

Each example includes:

- Business description  
- Valid curl  
- Sample ingestion response  
- Canonical JSON written to Raw S3 (with full compliance metadata)  

All compliance scenarios are exercised using the rules defined in `seed_compliance_rules.sql`.

---

# 1. X12 837 — Required Segments Present (PASS)

## Business Description
A clean 837 claim for member **123456789** and provider **123456789**.  
All X12 rules pass.

## Curl

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  --data '{
    "envelope": {
      "event_id": "evt-x12-001",
      "event_type": "837",
      "source_system": "claims-gateway",
      "produced_at": "2026-05-16T12:00:00Z",
      "format": "x12"
    },
    "payload": "ISA*00*          *00*          *ZZ*SENDERID       *ZZ*RECEIVERID     *260516*1200*^*00501*000000905*0*T*:~GS*HC*SENDERID*RECEIVERID*20260516*1200*1*X*005010X222A1~ST*837*0001*005010X222A1~NM1*IL*1*DOE*JOHN****MI*123456789~NM1*85*2*GOOD CLINIC*****XX*123456789~CLM*123456*100***11:B:1*Y*A*Y*Y~SE*7*0001~GE*1*1~IEA*1*000000905~"
  }' \
  http://hesp-dev-alb/events/ingest
```

## Ingestion Response

```json
{
  "event_id": "evt-x12-001",
  "ingested_at": "2026-05-16T12:01:00Z",
  "format": "x12"
}
```

## Raw S3 Canonical JSON (PASS)

```json
{
  "event_id": "evt-x12-001",
  "source_system": "claims-gateway",
  "format": "x12",
  "metadata": {
    "event_type": "837",
    "tenant_id": "acme-payer"
  },
  "patient": {
    "id": "123456789",
    "first_name": "JOHN",
    "last_name": "DOE"
  },
  "encounter": {
    "id": "123456",
    "type": "PROFESSIONAL_CLAIM"
  },
  "raw_value": "ISA*00*...~",
  "compliance_applied": true,
  "compliance_flag": true,
  "compliance_reason": "",
  "compliance_rule_type": "x12_837_required_segments_present",
  "compliance_rule_id": "123456789",
  "compliance_timestamp": "2026-05-16T12:01:00Z"
}
```

---

# 2. X12 837 — Missing Required Segment (FAIL)

## Business Description
Member **999999999** is configured to fail `x12_837_required_segments_present` with reason `MISSING_NM1_IL`.

## Curl

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  --data '{
    "envelope": {
      "event_id": "evt-x12-002",
      "event_type": "837",
      "source_system": "claims-gateway",
      "produced_at": "2026-05-16T12:05:00Z",
      "format": "x12"
    },
    "payload": "ISA*00*          *00*          *ZZ*SENDERID       *ZZ*RECEIVERID     *260516*1205*^*00501*000000906*0*T*:~GS*HC*SENDERID*RECEIVERID*20260516*1205*2*X*005010X222A1~ST*837*0002*005010X222A1~CLM*999999*200***11:B:1*Y*A*Y*Y~SE*4*0002~GE*1*2~IEA*1*000000906~"
  }' \
  http://hesp-dev-alb/events/ingest
```

## Ingestion Response

```json
{
  "event_id": "evt-x12-002",
  "ingested_at": "2026-05-16T12:05:30Z",
  "format": "x12"
}
```

## Raw S3 Canonical JSON (FAIL)

```json
{
  "event_id": "evt-x12-002",
  "source_system": "claims-gateway",
  "format": "x12",
  "metadata": {
    "event_type": "837",
    "tenant_id": "acme-payer"
  },
  "patient": {
    "id": "999999999"
  },
  "encounter": {
    "id": "999999",
    "type": "PROFESSIONAL_CLAIM"
  },
  "raw_value": "ISA*00*...~",
  "compliance_applied": true,
  "compliance_flag": false,
  "compliance_reason": "MISSING_NM1_IL",
  "compliance_rule_type": "x12_837_required_segments_present",
  "compliance_rule_id": "999999999",
  "compliance_timestamp": "2026-05-16T12:05:30Z"
}
```

---

# 3. HL7 ADT — Valid Encounter & Patient (PASS)

## Business Description
Encounter **123456** and patient **123456789** both have passing HL7 rules.

## Curl

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  --data '{
    "envelope": {
      "event_id": "evt-hl7-001",
      "event_type": "ADT",
      "source_system": "hospital-adt",
      "produced_at": "2026-05-16T12:10:00Z",
      "format": "hl7"
    },
    "payload": "MSH|^~\\&|ADT|HOSPITAL|HESP|INGEST|202605161210||ADT^A01|123456|P|2.5\rPID|1||123456789^^^HOSPITAL^MR||DOE^JOHN||19800101|M|||123 MAIN ST^^RALEIGH^NC^27612||5551234567\rPV1|1|I|ER||||P001^PRIMARYCARE|||||||||||123456"
  }' \
  http://hesp-dev-alb/events/ingest
```

## Ingestion Response

```json
{
  "event_id": "evt-hl7-001",
  "ingested_at": "2026-05-16T12:10:30Z",
  "format": "hl7"
}
```

## Raw S3 Canonical JSON (PASS)

```json
{
  "event_id": "evt-hl7-001",
  "source_system": "hospital-adt",
  "format": "hl7",
  "metadata": {
    "event_type": "ADT",
    "tenant_id": "acme-hospital"
  },
  "patient": {
    "id": "123456789",
    "first_name": "JOHN",
    "last_name": "DOE"
  },
  "encounter": {
    "id": "123456",
    "type": "INPATIENT"
  },
  "raw_value": "MSH|^~\\&|ADT|...",
  "compliance_applied": true,
  "compliance_flag": true,
  "compliance_reason": "",
  "compliance_rule_type": "hl7_pid_required",
  "compliance_rule_id": "123456",
  "compliance_timestamp": "2026-05-16T12:10:30Z"
}
```

---

# 4. FHIR Encounter — Valid Encounter & Patient (PASS)

## Business Description
FHIR Encounter with:

- Encounter ID **123456**
- Patient reference **987654**

Both rules pass.

## Curl

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  --data '{
    "envelope": {
      "event_id": "evt-fhir-001",
      "event_type": "Encounter",
      "source_system": "ehr-fhir",
      "produced_at": "2026-05-16T12:15:00Z",
      "format": "fhir"
    },
    "payload": {
      "resourceType": "Encounter",
      "id": "123456",
      "status": "in-progress",
      "subject": { "reference": "Patient/987654" }
    }
  }' \
  http://hesp-dev-alb/events/ingest
```

## Ingestion Response

```json
{
  "event_id": "evt-fhir-001",
  "ingested_at": "2026-05-16T12:15:30Z",
  "format": "fhir"
}
```

## Raw S3 Canonical JSON (PASS)

```json
{
  "event_id": "evt-fhir-001",
  "source_system": "ehr-fhir",
  "format": "fhir",
  "metadata": {
    "event_type": "Encounter",
    "tenant_id": "acme-provider"
  },
  "patient": {
    "id": "987654"
  },
  "encounter": {
    "id": "123456",
    "type": "ENCOUNTER"
  },
  "raw_value": {
    "resourceType": "Encounter",
    "id": "123456",
    "status": "in-progress",
    "subject": { "reference": "Patient/987654" }
  },
  "compliance_applied": true,
  "compliance_flag": true,
  "compliance_reason": "",
  "compliance_rule_type": "fhir_encounter_status_required",
  "compliance_rule_id": "123456",
  "compliance_timestamp": "2026-05-16T12:15:30Z"
}
```

---

# 5. Generic JSON — Valid Encounter & Member (PASS)

## Business Description
Generic JSON event with:

- Encounter ID **123456**
- Member ID **AET123987**

Both rules pass.

## Curl

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  --data '{
    "envelope": {
      "event_id": "evt-json-001",
      "event_type": "generic",
      "source_system": "partner-api",
      "produced_at": "2026-05-16T12:20:00Z",
      "format": "json"
    },
    "payload": {
      "encounter_id": "123456",
      "member_id": "AET123987",
      "status": "OPEN"
    }
  }' \
  http://hesp-dev-alb/events/ingest
```

## Ingestion Response

```json
{
  "event_id": "evt-json-001",
  "ingested_at": "2026-05-16T12:20:30Z",
  "format": "json"
}
```

## Raw S3 Canonical JSON (PASS)

```json
{
  "event_id": "evt-json-001",
  "source_system": "partner-api",
  "format": "json",
  "metadata": {
    "event_type": "generic",
    "tenant_id": "acme-partner"
  },
  "patient": {
    "id": "AET123987"
  },
  "encounter": {
    "id": "123456",
    "type": "GENERIC"
  },
  "raw_value": {
    "encounter_id": "123456",
    "member_id": "AET123987",
    "status": "OPEN"
  },
  "compliance_applied": true,
  "compliance_flag": true,
  "compliance_reason": "",
  "compliance_rule_type": "generic_encounter_id_required",
  "compliance_rule_id": "123456",
  "compliance_timestamp": "2026-05-16T12:20:30Z"
}
```

---

# Summary of Rules Exercised

| Format | Rule                              | Entity             | Result |
|--------|-----------------------------------|--------------------|--------|
| X12    | x12_837_required_segments_present | member 123456789   | PASS   |
| X12    | x12_837_required_segments_present | member 999999999   | FAIL   |
| X12    | billing_provider_npi_valid        | provider 123456789 | PASS   |
| X12    | x12_encounter_requires_clm        | encounter 123456   | PASS   |
| HL7    | hl7_pid_required                  | encounter 123456   | PASS   |
| HL7    | hl7_patient_id_valid              | patient 123456789  | PASS   |
| FHIR   | fhir_encounter_status_required    | encounter 123456   | PASS   |
| FHIR   | fhir_patient_reference_valid      | patient 987654     | PASS   |
| JSON   | generic_encounter_id_required     | encounter 123456   | PASS   |
| JSON   | generic_member_id_valid           | member AET123987   | PASS   |
```