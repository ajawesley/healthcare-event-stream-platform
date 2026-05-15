---

# Healthcare Event Stream Platform — IaC Architecture

## 1. Purpose
The Infrastructure‑as‑Code (IaC) architecture defines how the Healthcare Event Stream Platform (HESP) is provisioned, governed, and deployed. It provides a modular, reusable Terraform foundation that enables teams to onboard new ingestion workloads, compliance services, and data pipelines using consistent patterns and guardrails.

The IaC model ensures that every environment is secure, reproducible, and aligned with enterprise governance and HIPAA requirements.

---

## 2. Terraform Architecture Overview
The IaC architecture is organized into four major module groups:

### **2.1 Core Landing Zone Modules**
These modules establish the secure network and encryption foundation:

- VPC (public, private, isolated subnets)  
- Route tables and NAT gateways  
- Security groups  
- VPC endpoints (S3, STS, SSM, ECR, Logs, Secrets Manager)  
- KMS keys for encryption  

### **2.2 Governance & Security Modules**
These modules enforce enterprise and HIPAA‑aligned governance:

- AWS Config  
- CloudTrail  
- Security Hub  
- GuardDuty  
- Inspector  
- Service Control Policies (SCPs)  
- IAM permission boundaries  
- Log archive S3 bucket  

### **2.3 Data Plane Modules**
These modules provide durable, compliant storage:

- Raw S3 bucket (unmodified payloads)  
- Curated S3 bucket (normalized and enriched datasets)  
- RDS PostgreSQL (compliance metadata)  
- DynamoDB (rules and configuration)  
- Redis (low‑latency rule caching)  

### **2.4 Application Runtime Modules**
These modules deploy the ingestion and compliance services:

- ECS cluster (Fargate)  
- ECS services (ingestion, compliance engine, ADOT collector)  
- IAM execution roles  
- IAM task roles  

---

## 3. Module Responsibilities

### **3.1 Core Landing Zone Modules**
Provide the foundational network and encryption controls:

- segmented subnets for PHI isolation  
- least‑privilege security groups  
- private access to AWS services  
- KMS‑encrypted data stores  

These modules ensure PHI never leaves the controlled network boundary.

---

### **3.2 Governance & Security Modules**
Provide continuous compliance and auditability:

- configuration drift detection  
- immutable audit logs  
- consolidated security posture  
- threat detection  
- vulnerability scanning  
- enforced least privilege  
- long‑term log retention  

These controls apply uniformly across all workloads.

---

### **3.3 Data Plane Modules**
Provide durable, compliant, replay‑safe storage:

- raw payload retention  
- curated datasets for downstream consumers  
- lifecycle and compliance metadata  
- rule evaluation and caching  

All data stores are encrypted, isolated, and access‑controlled.

---

### **3.4 Application Runtime Modules**
Provide the compute and runtime environment:

- ingestion service  
- compliance engine  
- ADOT collector for traces and metrics  
- IAM roles for scoped access  

The runtime layer is fully observable and supports safe deployment patterns.

---

## 4. CI/CD Workflow & Deployment Safety

### **4.1 Workflow Overview**
The CI/CD pipeline performs:

1. static validation (fmt, validate, lint)  
2. Terraform plan with drift detection  
3. policy checks (SCP + permission boundaries)  
4. controlled apply  
5. post‑deployment verification (health checks + SLO checks)  

All changes are version‑controlled and auditable.

---

### **4.2 Safe Deployment Controls**
The platform enforces:

- rolling or blue/green deployments  
- ALB health‑check gates  
- dependency health checks (RDS, Redis, DynamoDB, S3)  
- trace and metric‑based verification  
- canary rollout stages  

Deployments cannot progress if any gate fails.

---

### **4.3 Automated Rollback**
Rollback is triggered automatically when:

- error rates exceed thresholds  
- ingestion or compliance failures spike  
- S3 or RDS latency violates SLOs  
- health checks fail during rollout  
- dependency availability degrades  

Rollback protects PHI integrity and minimizes blast radius.

---

### **4.4 Immutable Infrastructure**
All infrastructure is:

- declarative  
- versioned  
- reproducible  
- environment‑scoped  

No manual changes are permitted in production environments.

---

## 5. Developer Reuse & Extensibility

### **5.1 Reusable Modules**
Teams can reuse modules for:

- new ingestion services  
- new compliance engines  
- new data pipelines  
- new S3 landing zones  
- new RDS or DynamoDB stores  

Modules enforce consistent patterns and governance.

---

### **5.2 Environment Parity**
The same IaC stack deploys:

- dev  
- staging  
- production  

This ensures consistent behavior across environments.

---

### **5.3 Extensible Architecture**
New capabilities can be added by:

- composing existing modules  
- creating new modules following platform patterns  
- extending the data plane  
- adding new event types or workflows  

The IaC model supports long‑term platform evolution.

---

## 6. Outcomes

- secure, governed, HIPAA‑aligned infrastructure  
- reusable Terraform modules for rapid onboarding  
- safe, observable, rollback‑enabled deployments  
- consistent environments across the enterprise  
- strong PHI boundaries and auditability  
- scalable foundation for future healthcare workloads  

---