# AcmeCo Event Stream Platform — Consumer Integration Guide

**Platform:** AcmeCo Event Stream Platform (ESP)  
**Audience:** Engineering teams consuming events from ESP  
**Status:** Canonical reference — do not modify without Platform Architecture review  
**Last reviewed:** 2026‑05

---

## Overview

This guide provides engineering teams with the requirements, responsibilities, and best practices for consuming events from the AcmeCo Event Stream Platform (ESP). It covers:

- consumer responsibilities  
- stream access and ACLs  
- idempotent processing  
- PHI/PII handling  
- offset management  
- replay safety  
- compliance metadata usage  
- testing and operational readiness  

This guide is mandatory for all teams onboarding as consumers.

---

## Consumer Responsibilities

Consumers are responsible for safely and correctly processing events from ESP streams. Consumers must:

- honor `data_classification`  
- store PHI only in approved systems  
- implement idempotent processing keyed on `event_id`  
- acknowledge events only after durable processing  
- handle replay safely  
- support offset resets  
- respect tenant boundaries  
- read and apply compliance metadata  

Consumers **must not**:

- write PHI to non‑approved systems  
- exfiltrate PHI outside AcmeCo networks  
- assume exactly‑once delivery  
- ignore compliance metadata  
- bypass stream ACLs  
- log PHI or PII  

---

## Stream Access & ACLs

Access to ESP streams is controlled by:

- tenant‑scoped ACLs  
- classification‑scoped ACLs  
- event‑type‑scoped ACLs  

### **Access Requirements**

To consume from a stream, teams must:

1. Request access through the ESP Access service  
2. Provide justification and intended use  
3. Undergo compliance review (for PHI/PII)  
4. Receive explicit approval  

No wildcard access is permitted for PHI or PII streams.

---

## Event Consumption Model

ESP provides:

- **at‑least‑once delivery**  
- **partition‑level ordering**  
- **offset tracking**  
- **replay support**  

Consumers must be designed to handle:

- duplicates  
- out‑of‑order events across partitions  
- replayed historical events  
- offset resets  

---

## Idempotent Processing

Idempotency is mandatory.

Consumers must use:

```
event_id
```

as the idempotency key.

### **Idempotency Strategies**

- store processed `event_id`s in a durable store  
- use upsert semantics  
- use natural keys in downstream systems  
- avoid side effects before idempotency checks  

### **Non‑Idempotent Patterns (Prohibited)**

- incrementing counters  
- appending logs without dedupe  
- triggering external calls before validation  

---

## Compliance Metadata Usage

Consumers must read and respect compliance metadata:

| Field | Meaning |
|---|---|
| `compliance_flag` | Whether any rule fired |
| `compliance_reason` | Explanation of rule outcome |
| `compliance_rule_type` | Category of rule |
| `compliance_rule_id` | Identifier of rule |
| `compliance_timestamp` | When evaluation occurred |

### **Consumer Responsibilities**

- enforce retention overrides  
- apply masking rules if required  
- restrict routing based on compliance outcomes  
- avoid processing events flagged for violations  

---

## PHI/PII Handling Requirements

Consumers must:

- store PHI only in approved systems  
- encrypt PHI at rest  
- restrict access to authorized users  
- avoid logging PHI  
- avoid exporting PHI to non‑approved destinations  

### **Prohibited Actions**

- writing PHI to logs  
- sending PHI to analytics tools without approval  
- storing PHI in caches without encryption  
- exposing PHI in metrics or traces  

---

## Offset Management

Consumers must manage offsets safely.

### **Rules**

- commit offsets only after durable processing  
- do not commit offsets speculatively  
- support offset resets  
- support replay scenarios  

### **Offset Reset Use Cases**

- consumer bug fix  
- downstream outage recovery  
- schema migration  
- replay‑driven reprocessing  

---

## Replay Safety

Consumers must be replay‑safe.

Replay may:

- re‑emit historical events  
- re‑emit events with updated compliance metadata  
- re‑emit events in large batches  
- re‑emit events out of sync with real‑time traffic  

### **Replay‑Safe Requirements**

- idempotent processing  
- no reliance on wall‑clock time  
- no assumptions about event uniqueness  
- no irreversible side effects before validation  

---

## Consumer Testing

### **Local Testing**

Consumers must validate:

- idempotency  
- offset handling  
- compliance metadata handling  
- PHI masking logic  
- replay simulation  

### **Integration Testing**

Integration tests validate:

- stream access  
- offset reset behavior  
- replay behavior  
- compliance metadata propagation  
- PHI storage validation  

---

## Operational Readiness

Before production approval, consumers must pass the ESP Operational Readiness Review (ORR).

### **ORR Checklist**

| Requirement | Status |
|---|---|
| Idempotency implemented | ☐ |
| PHI storage approved | ☐ |
| Offset reset supported | ☐ |
| Replay‑safe logic | ☐ |
| Compliance metadata handled | ☐ |
| Dashboards created | ☐ |
| Alerts configured | ☐ |
| Runbooks created | ☐ |
| On‑call rotation established | ☐ |

---

## Common Pitfalls

New consumers often encounter:

- ignoring compliance metadata  
- logging PHI  
- committing offsets too early  
- failing idempotency checks  
- breaking on replay  
- assuming exactly‑once delivery  
- storing PHI in non‑approved systems  

These issues are caught during onboarding reviews.

---

## Summary

This guide defines the requirements for consuming events from the AcmeCo Event Stream Platform. By following these standards, consumers ensure:

- safe handling of PHI/PII  
- correct application of compliance metadata  
- reliable, idempotent processing  
- replay‑safe behavior  
- secure and compliant downstream storage  

Consumers play a critical role in maintaining the integrity and compliance posture of the platform.
