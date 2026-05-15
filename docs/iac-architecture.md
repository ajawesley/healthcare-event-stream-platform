Healthcare Event Stream Platform — IaC Architecture
1. Purpose
The Infrastructure‑as‑Code (IaC) architecture defines how the Healthcare Event Stream Platform (HESP) is provisioned, governed, and deployed. It provides a modular, reusable Terraform foundation that enables teams to onboard new ingestion workloads, compliance services, and data pipelines using consistent patterns and guardrails.

The IaC model ensures that every environment is secure, reproducible, and aligned with enterprise governance and HIPAA requirements.

2. Terraform Architecture Overview
<pre>

mermaid
flowchart TD
    ROOT[Root Stack<br/>infra/envs/dev]

    subgraph Core[Core Landing Zone Modules]
        VPC[VPC Module<br/>VPC, Subnets, Routes, NAT]
        SG[Security Groups Module]
        ENDPOINTS[VPC Endpoints Module<br/>S3, STS, SSM, ECR, Logs, Secrets]
        KMS[KMS Module<br/>Encryption Keys]
    end

    subgraph Governance[Governance & Security Modules]
        CONFIG[AWS Config Module]
        CLOUDTRAIL[CloudTrail Module]
        HUB[Security Hub Module]
        GD[GuardDuty Module]
        INSPECTOR[Inspector Module]
        SCP[SCPs]
        PB[Permission Boundaries]
        LOG_ARCHIVE[Log Archive S3 Module]
    end

    subgraph Data[Data Plane Modules]
        S3_RAW[Raw Events S3 Module]
        S3_CURATED[Curated / Golden S3 Module]
        RDS[RDS PostgreSQL Module<br/>Compliance DB]
        DDB[DynamoDB Module<br/>Compliance Rules]
        REDIS[ElastiCache Redis Module]
    end

    subgraph App[Application Runtime Modules]
        ECS_CLUSTER[ECS Cluster Module]
        ECS_SERVICE[ECS Service Module<br/>Ingest + Compliance + ADOT]
        IAM_EXEC[IAM Execution Role Module]
        IAM_TASK[IAM Task Role Module]
    end

    ROOT --> Core
    ROOT --> Governance
    ROOT --> Data
    ROOT --> App

    Core --> App
    Core --> Data
    Core --> Governance

    App --> S3_RAW
    App --> S3_CURATED
    App --> RDS
    App --> DDB
    App --> REDIS

    Governance --> LOG_ARCHIVE
    CLOUDTRAIL --> LOG_ARCHIVE
    CONFIG --> LOG_ARCHIVE
</pre>

3. Module Responsibilities
3.1 Core Landing Zone Modules
These modules establish the secure network and encryption foundation:

VPC with public, private, and isolated subnets

Security groups enforcing least‑privilege network boundaries

VPC endpoints for private access to AWS services

KMS keys for S3, RDS, and application secrets

These modules ensure PHI never leaves the controlled network boundary.

3.2 Governance & Security Modules
These modules enforce enterprise and HIPAA‑aligned governance:

AWS Config for configuration compliance

CloudTrail for immutable audit logs

Security Hub for consolidated security posture

GuardDuty for threat detection

Inspector for vulnerability scanning

SCPs restricting unsafe operations

IAM permission boundaries enforcing least privilege

Log archive bucket for immutable long‑term storage

Governance modules apply uniformly across all workloads.

3.3 Data Plane Modules
These modules provide durable, compliant storage:

Raw S3 bucket for unmodified payloads

Curated S3 bucket for normalized and enriched datasets

RDS PostgreSQL for compliance metadata and lifecycle state

DynamoDB for rules and configuration

Redis for low‑latency rule caching

All data stores are encrypted, isolated, and access‑controlled.

3.4 Application Runtime Modules
These modules deploy the ingestion and compliance services:

ECS cluster for Fargate compute

ECS service for ingestion, compliance engine, and ADOT sidecar

IAM execution role for ECR pulls, logging, and metrics

IAM task role for S3, RDS, DynamoDB, Redis, and Secrets Manager

The application layer is fully observable and supports safe deployment patterns.

4. CI/CD Workflow & Deployment Safety
4.1 Workflow Overview
The CI/CD pipeline performs:

Static validation (fmt, validate, lint)

Terraform plan with drift detection

Policy checks (SCP + permission boundaries)

Controlled apply

Post‑deployment verification (health checks + SLO checks)

All changes are version‑controlled and auditable.

4.2 Safe Deployment Controls
The platform enforces:

rolling or blue/green deployments

ALB health‑check gates

dependency health checks (RDS, Redis, DynamoDB, S3)

trace and metric‑based verification

canary rollout stages

Deployments cannot progress if any gate fails.

4.3 Automated Rollback
Rollback is triggered automatically when:

error rates exceed thresholds

ingestion or compliance failures spike

S3 or RDS latency violates SLOs

health checks fail during rollout

dependency availability degrades

Rollback protects PHI integrity and minimizes blast radius.

4.4 Immutable Infrastructure
All infrastructure is:

declarative

versioned

reproducible

environment‑scoped

No manual changes are permitted in production environments.

5. Developer Reuse & Extensibility
5.1 Reusable Modules
Teams can reuse modules for:

new ingestion services

new compliance engines

new data pipelines

new S3 landing zones

new RDS or DynamoDB stores

Modules enforce consistent patterns and governance.

5.2 Environment Parity
The same IaC stack deploys:

dev

staging

production

This ensures consistent behavior across environments.

5.3 Extensible Architecture
New capabilities can be added by:

composing existing modules

creating new modules following platform patterns

extending the data plane

adding new event types or workflows

The IaC model supports long‑term platform evolution.

6. Outcomes
Secure, governed, HIPAA‑aligned infrastructure

Reusable Terraform modules for rapid onboarding

Safe, observable, rollback‑enabled deployments

Consistent environments across the enterprise

Strong PHI boundaries and auditability

A scalable foundation for future healthcare workloads