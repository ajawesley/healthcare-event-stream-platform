# IaC Guardrails & Paved Roads — Terraform Implementation Guide

This document provides **Terraform guardrail snippets**, **paved‑road examples**, and **module‑level enforcement patterns** for the Healthcare Event Stream Platform (HESP).  
It complements the `iac-architecture.md` by showing **how the guardrails are implemented in code**.

---

# 1. Purpose

This guide ensures:

- consistent Terraform patterns  
- HIPAA‑aligned guardrails  
- paved‑road defaults  
- safe, reproducible infrastructure  
- no “snowflake” resources  

Every module includes:

- **Guardrails** → enforced constraints  
- **Paved Roads** → recommended defaults  
- **Terraform Snippets** → copy‑ready examples  

---

# 2. Core Landing Zone Guardrails

---

## 2.1 VPC Module Guardrails

### Guardrails
- No public IPs  
- Private subnets required  
- Flow logs mandatory  
- NAT gateways only for private egress  
- Default SG locked down  

### Paved Road
All workloads deploy into **private subnets** and use **VPC endpoints**.

### Terraform Snippet — VPC with Guardrails

```hcl
module "vpc" {
  source = "../modules/vpc"

  name = "hesp-dev"

  cidr = "10.10.0.0/16"

  enable_dns_hostnames = true
  enable_dns_support   = true

  public_subnets  = []
  private_subnets = ["10.10.1.0/24", "10.10.2.0/24"]
  isolated_subnets = ["10.10.3.0/24"]

  enable_flow_logs = true

  map_public_ip_on_launch = false

  tags = {
    "phi" = "isolated"
    "env" = "dev"
  }
}
