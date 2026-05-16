# Healthcare Event Stream Platform — Landing Zone Architecture

## 1. Purpose
The landing zone provides the secure, governed, and compliant foundation on which the Healthcare Event Stream Platform (HESP) operates. It establishes the network, identity, security, and observability controls required to process PHI safely and to support clinical, operational, and administrative healthcare workflows.

The landing zone is designed to be healthcare‑ready, HIPAA‑aligned, and extensible across multiple business domains.

---

## 2. Architecture Overview

```mermaid
flowchart LR
    %% External Sources
    subgraph External[External Healthcare Sources]
        HL7[HL7 v2 Feeds]
        X12[X12 EDI Transactions]
        FHIR[FHIR APIs]
        REST[Partner REST Integrations]
    end

    %% Edge Layer
    subgraph Edge[Ingress & Security Controls]
        WAF[WAF / Security Policies]
        ALB[Application Load Balancer<br/>TLS Termination]
    end

    %% Application Layer
    subgraph App[ECS Application Layer]
        INGEST[Ingestion Service (Go)]
        COMPLIANCE[Compliance Engine (Go)]
        ADOT[ADOT Collector]
    end

    %% Data Plane
    subgraph Data[Data Plane]
        RAW_S3[(Raw Events S3 Bucket)]
        CURATED_S3[(Curated / Golden S3 Bucket)]
        RDS[(RDS PostgreSQL<br/>Compliance DB)]
        DDB[(DynamoDB<br/>Compliance Rules)]
        REDIS[(ElastiCache Redis<br/>Low‑Latency Cache)]
    end

    %% Observability
    subgraph Obs[Observability]
        CWL[CloudWatch Logs]
        CWM[CloudWatch Metrics]
        XRAY[X-Ray Traces]
        LOG_ARCHIVE[(Log Archive S3)]
    end

    %% Governance
    subgraph Gov[Governance & Security]
        CONFIG[AWS Config]
        CLOUDTRAIL[CloudTrail]
        HUB[Security Hub]
        GD[GuardDuty]
        INSPECTOR[Inspector]
        SCP[SCPs]
        PB[Permission Boundaries]
    end

    %% Network
    subgraph Network[VPC Landing Zone]
        PUB[Public Subnets<br/>(ALB)]
        PRIV[Private Subnets<br/>(ECS)]
        ISO[Isolated Subnets<br/>(RDS, Redis)]
        VPCE[VPC Endpoints<br/>S3, STS, SSM, ECR, Logs, Secrets]
    end

    %% Flows
    HL7 --> WAF
    X12 --> WAF
    FHIR --> WAF
    REST --> WAF
    WAF --> ALB
    ALB --> INGEST

    INGEST --> RAW_S3
    INGEST --> COMPLIANCE

    COMPLIANCE --> RDS
    COMPLIANCE --> DDB
    COMPLIANCE --> REDIS
    COMPLIANCE --> CURATED_S3

    INGEST --> CWL
    COMPLIANCE --> CWL
    ADOT --> XRAY
    ADOT --> CWM

    CWL --> LOG_ARCHIVE
    CLOUDTRAIL --> LOG_ARCHIVE
    CONFIG --> LOG_ARCHIVE

    Network --- Edge
    Network --- App
    Network --- Data
    Network --- Obs
    Network --- Gov
    VPCE --- App
    VPCE --- Data
