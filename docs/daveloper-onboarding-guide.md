# AcmeCo Event Stream Platform — Developer Onboarding Guide

**Platform:** AcmeCo Event Stream Platform (ESP)  
**Audience:** New AcmeCo developers integrating with or contributing to ESP  
**Status:** Canonical reference — do not modify without Platform Architecture review  
**Last reviewed:** 2026‑05

---

## Overview

This guide provides new developers with the essential knowledge required to build, test, and operate services on the AcmeCo Event Stream Platform (ESP). It covers:

- platform concepts  
- development environment setup  
- producer and consumer responsibilities  
- event model requirements  
- compliance and security expectations  
- testing and validation workflows  
- operational readiness  
- platform‑as‑a‑product patterns and reusable components

This guide is required reading for all engineers onboarding to ESP.

---

## Platform Concepts

Before writing code, developers must understand the core components of ESP:

### **1. Envelope + Payload Model**
All events must follow the canonical envelope and payload structure defined in the Event Model.

### **2. Canonicalization**
Raw events are normalized into canonical structures (patient, encounter, observation).

### **3. Compliance Engine**
Every event is evaluated against compliance rules and annotated with compliance metadata.

### **4. Stream Routing**
Events are routed to tenant‑scoped, classification‑scoped topics.

### **5. Replay**
Historical events can be reprocessed through controlled, auditable replay workflows.

### **6. Retention & Archival**
Events are retained for 90 days on the active stream and 1–7 years in the archive.

---

# Platform as a Product  
### **How to Reuse Platform Patterns Instead of Re‑Inventing Them**

ESP is built as a **product**, not a collection of scripts.  
Teams must **consume** platform capabilities rather than rebuild them.

Below are the reusable patterns, libraries, modules, and workflows that every team must adopt.

---

# 🔧 Go Application Patterns (Shared Libraries)

ESP provides a set of Go modules that standardize service behavior across the platform.

### **1. Probes Framework (Liveness + Readiness)**  
All Go services must use the shared `probes` package.

Benefits:
- consistent `/healthz` and `/readyz` endpoints  
- pluggable readiness checks (DB, Kafka, S3, schema registry)  
- automatic integration with Kubernetes  
- zero custom boilerplate  

Example readiness extension:
```go
probes.RegisterReadinessCheck("kafka", kafkaClient.Ready)
```

### **2. Resiliency Module**
All services must use the shared resiliency library, which provides:

- exponential backoff  
- circuit breakers  
- retry policies  
- timeouts  
- context propagation  
- structured error wrapping  

### **3. Logging + Tracing**
All services must use the platform logging module:

- structured JSON logs  
- correlation IDs  
- event_id propagation  
- OpenTelemetry tracing  
- PHI‑safe logging enforcement  

### **4. Configuration Pattern**
All services use the shared config loader:

- environment‑first  
- typed config structs  
- validation on startup  
- secrets pulled from AWS Secrets Manager  

---

# 🏗 Terraform Patterns (Infrastructure as Product)

ESP provides a set of Terraform modules that developers must reuse:

### **1. VPC Module**
- public, private, isolated subnets  
- NAT gateways per AZ  
- VPC endpoints  
- route tables  
- tagging standards  

### **2. ECS Service Module**
- task definitions  
- autoscaling  
- service mesh integration  
- health checks  
- logging + X‑Ray  
- IAM least‑privilege roles  

### **3. Lambda Module**
- runtime configuration  
- DLQ + retry policies  
- IAM roles  
- environment variable injection  
- VPC attachment  

### **4. Stream Infrastructure Module**
- Kafka topics  
- ACLs  
- consumer groups  
- retention policies  
- encryption  

### **5. Compliance Engine Module**
- rule evaluation Lambda  
- DynamoDB rule store  
- S3 compliance logs  
- CloudWatch dashboards  

### **6. Governance Module**
- tagging enforcement  
- IAM boundary policies  
- resource naming conventions  
- mandatory CloudWatch alarms  

**Do not create infrastructure manually.  
Do not fork platform modules.  
Always consume the official modules.**

---

# 🚀 CI/CD Workflow Patterns

All services must use the standard ESP CI/CD workflow:

### **Pipeline Stages**

1. **Static Analysis**
   - `golangci-lint`
   - dependency vulnerability scan  
   - PHI‑safe logging scan  

2. **Unit Tests**
   - coverage threshold enforced  
   - mocks for external systems  

3. **Schema Validation**
   - producer schemas validated against registry  
   - consumer schemas validated for compatibility  

4. **Integration Tests**
   - ingestion mock  
   - compliance mock  
   - replay simulation  

5. **Security & Governance**
   - OPA policy checks  
   - Terraform plan validation  
   - IAM least‑privilege scan  

6. **Build + Artifact Publish**
   - container image  
   - SBOM  
   - provenance metadata  

7. **Deploy to Shared Services**
   - ephemeral environment  
   - smoke tests  
   - rollback hooks  

8. **Promote to Production**
   - manual approval  
   - compliance sign‑off  
   - automated canary rollout  

---

# 🛡 Governance Patterns

ESP enforces governance through:

### **1. Tagging Standards**
Every resource must include:

- `app_name`  
- `owner`  
- `data_classification`  
- `environment`  
- `cost_center`  

### **2. IAM Boundaries**
All roles are constrained by:

- least‑privilege policies  
- PHI‑safe access boundaries  
- no wildcard permissions  

### **3. Compliance Enforcement**
- PHI encryption required  
- classification required  
- compliance metadata required  
- retention enforced automatically  

### **4. Observability Baseline**
Every service must emit:

- structured logs  
- metrics (latency, throughput, errors)  
- traces  
- health probes  

### **5. Operational Readiness**
No service enters production without:

- dashboards  
- alerts  
- runbooks  
- on‑call rotation  
- replay impact analysis  

---

# Development Environment Setup

### **Prerequisites**

- AWS CLI (latest)  
- Docker  
- Go 1.22+  
- Terraform 1.6+  
- Make  
- jq  
- Access to the ESP GitHub repositories  
- Access to the ESP Shared Services AWS account  

### **Local Development Tools**

ESP provides:

- local ingestion mock  
- schema registry mock  
- compliance engine mock  
- replay simulator  
- event generator CLI  

These tools allow developers to test producers and consumers without deploying to AWS.

---

# Producer Onboarding

Producers are responsible for constructing valid events and submitting them to the ingestion API.

### **Producer Responsibilities**

- Generate UUID v4 `event_id`  
- Set `produced_at` in UTC  
- Encrypt PHI/PII fields in the payload  
- Populate all required envelope fields  
- Register `source_system` in the Service Registry  
- Validate payload schema against the Schema Registry  
- Handle ingestion retries safely using `event_id`  

### **Producer Must Not**

- set `ingested_at`  
- send unencrypted PHI  
- misclassify PHI as `internal` or `public`  
- bypass the ingestion API  
- embed PHI in metadata fields  

### **Producer Checklist**

| Requirement | Status |
|---|---|
| Envelope fields populated | ☐ |
| Payload schema registered | ☐ |
| PHI encrypted | ☐ |
| Source system registered | ☐ |
| Ingestion retries implemented | ☐ |
| Local tests passing | ☐ |

---

# Consumer Onboarding

Consumers read events from ESP streams and process them according to business logic.

### **Consumer Responsibilities**

- honor `data_classification`  
- store PHI only in approved systems  
- implement idempotent processing keyed on `event_id`  
- acknowledge events only after durable processing  
- handle replay safely  
- support offset resets  

### **Consumer Must Not**

- write PHI to non‑approved stores  
- exfiltrate PHI  
- ignore compliance metadata  
- assume exactly‑once delivery  

### **Consumer Checklist**

| Requirement | Status |
|---|---|
| Idempotent processing implemented | ☐ |
| PHI storage approved | ☐ |
| Offset reset supported | ☐ |
| Replay‑safe logic | ☐ |
| Compliance metadata handled | ☐ |
| Local consumer tests passing | ☐ |

---

# Event Schema Registration

All payload schemas must be registered in the ESP Schema Registry before production deployment.

### **Schema Requirements**

- semantic versioning  
- backward‑compatible minor versions  
- breaking changes require major version bump  
- schema linked to `event_type` + `event_version`  
- automated CI validation  

### **Schema Registration Workflow**

1. Create or update schema  
2. Run local schema validator  
3. Submit PR to Schema Registry repo  
4. Automated checks validate compatibility  
5. Platform Architecture approves  
6. Schema is published  
7. Producer CI pipeline enforces schema validation  

---

# Compliance Alignment for Developers

Compliance is not optional. Developers must understand:

### **1. Classification**
Every event must declare `data_classification`.

### **2. PHI Handling**
PHI must be encrypted at the field level before ingestion.

### **3. Compliance Metadata**
Consumers must read and respect:

- `compliance_flag`  
- `compliance_reason`  
- `compliance_rule_type`  
- `compliance_rule_id`  

### **4. Retention**
Retention is enforced automatically based on classification and compliance rules.

### **5. Auditability**
All access is logged and retained for 7 years.

---

# Testing & Validation

### **Local Testing**

Developers must run:

- envelope validation  
- schema validation  
- compliance mock evaluation  
- ingestion mock tests  
- consumer idempotency tests  
- replay simulation tests  

### **Integration Testing**

Integration tests run in the ESP Shared Services environment:

- ingestion → canonicalization → compliance → routing  
- consumer offset reset  
- replay workflows  
- schema registry integration  

### **Pre‑Production Testing**

Before deployment:

- load tests  
- PHI masking validation  
- retention policy validation  
- ACL validation  
- compliance rule evaluation  

---

# Operational Readiness

Before a service is approved for production, it must pass the ESP Operational Readiness Review (ORR).

### **ORR Checklist**

| Requirement | Status |
|---|---|
| Runbooks created | ☐ |
| Dashboards created | ☐ |
| Alerts configured | ☐ |
| On‑call rotation established | ☐ |
| Replay impact assessed | ☐ |
| Compliance alignment validated | ☐ |
| Security review completed | ☐ |
| DR strategy documented | ☐ |

---

# Common Pitfalls

New developers often encounter:

- forgetting to encrypt PHI  
- misclassifying events  
- failing schema validation  
- non‑idempotent consumers  
- ignoring compliance metadata  
- writing PHI to logs  
- assuming exactly‑once delivery  
- failing to handle replay  

These issues are caught during onboarding reviews.

---

# Support & Resources

### **Slack Channels**

- `#esp-platform` — general questions  
- `#esp-producers` — producer onboarding  
- `#esp-consumers` — consumer onboarding  
- `#esp-compliance` — compliance rule questions  
- `#esp-ops` — operational support  

### **Documentation**

- Event Model  
- Lifecycle Model  
- Compliance Alignment  
- HIPAA Alignment  
- Security & Governance Baseline  
- Replay Model  

### **Training**

- ESP 101 (required)  
- PHI Handling Certification (required)  
- Replay Operations Training (optional)  
- Compliance Rule Authoring (optional)  

---

# Summary

This guide provides the foundational knowledge required for developers to build safe, compliant, and reliable integrations with the AcmeCo Event Stream Platform. By following the onboarding steps, validation workflows, and operational readiness requirements—and by reusing platform‑as‑a‑product patterns—developers ensure that all services meet AcmeCo’s security, compliance, and reliability standards.
