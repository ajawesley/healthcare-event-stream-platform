# Lifecycle Model

The lifecycle model defines the valid states and transitions for major healthcare workflows. HESP assigns lifecycle states during normalization to provide consistent semantics across heterogeneous systems.

## 1. Encounter Lifecycle

```
created → in_progress → finalized → amended → closed
```

### State Definitions
- **created** — encounter record created by EHR or intake system  
- **in_progress** — patient is being seen or encounter is actively updated  
- **finalized** — clinical documentation completed  
- **amended** — corrections or updates applied post-finalization  
- **closed** — encounter archived or no longer active  

---

## 2. Claim Lifecycle

```
submitted → received → adjudicating → pending_info → approved | denied → closed
```

### State Definitions
- **submitted** — provider submits claim  
- **received** — payer acknowledges receipt  
- **adjudicating** — claim under review  
- **pending_info** — additional documentation required  
- **approved** — claim approved for payment  
- **denied** — claim denied  
- **closed** — claim finalized  

---

## 3. Prior Authorization Lifecycle

```
requested → received → clinical_review → pending_info → approved | denied → appealed → finalized
```

---

## 4. Lab Order Lifecycle

```
ordered → collected → in_lab → resulted → corrected → finalized
```

---

## 5. Lifecycle Assignment Rules

- Lifecycle is determined by:
  - event type  
  - source system  
  - semantic mapping rules  
  - sequence ordering  
- Invalid transitions are logged and rejected.
- Lifecycle transitions are deterministic and replay-safe.

---

## 6. Design Principles

- **Consistency**: All workflows follow predictable state machines.
- **Traceability**: Every transition is auditable.
- **Extensibility**: New workflows can be added without breaking existing ones.
- **Compliance**: Lifecycle metadata supports audit, retention, and PHI governance.

