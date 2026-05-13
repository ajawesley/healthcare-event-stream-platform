-- seed_compliance_rules.sql

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS compliance_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type   TEXT NOT NULL,
    entity_id     TEXT NOT NULL,
    rule_type     TEXT NOT NULL,
    compliance_flag BOOLEAN NOT NULL,
    reason_code   TEXT,
    source_format TEXT NOT NULL,
    event_type    TEXT NOT NULL,
    created_at    TIMESTAMPTZ DEFAULT now(),
    updated_at    TIMESTAMPTZ DEFAULT now()
);

-- Ensure idempotency: one logical rule per (entity_type, entity_id, rule_type, source_format, event_type)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM   pg_constraint
        WHERE  conname = 'unique_rule'
        AND    conrelid = 'compliance_rules'::regclass
    ) THEN
        ALTER TABLE compliance_rules
        ADD CONSTRAINT unique_rule
        UNIQUE (entity_type, entity_id, rule_type, source_format, event_type);
    END IF;
END$$;

-- Helper comment:
-- All inserts below use UPSERT so re-running this file will UPDATE existing rules instead of duplicating them.

-- ---------------------------------------------------------
-- X12 Rules (original + encounter-based)
-- ---------------------------------------------------------

INSERT INTO compliance_rules (
    entity_type, entity_id, rule_type, compliance_flag,
    reason_code, source_format, event_type
) VALUES
    -- Original rules
    ('member',   '123456789', 'x12_837_required_segments_present', true,  NULL,           'x12', '837'),
    ('provider', '123456789', 'billing_provider_npi_valid',        true,  NULL,           'x12', '837'),
    ('member',   '999999999', 'x12_837_required_segments_present', false, 'MISSING_NM1_IL','x12', '837'),

    -- Encounter-based rule for demo/testing
    ('encounter','123456',    'x12_encounter_requires_clm',        true,  NULL,           'x12', '837')
ON CONFLICT (entity_type, entity_id, rule_type, source_format, event_type)
DO UPDATE SET
    compliance_flag = EXCLUDED.compliance_flag,
    reason_code     = EXCLUDED.reason_code,
    updated_at      = now();

-- ---------------------------------------------------------
-- HL7 Rules
-- ---------------------------------------------------------

INSERT INTO compliance_rules (
    entity_type, entity_id, rule_type, compliance_flag,
    reason_code, source_format, event_type
) VALUES
    -- Encounter-level HL7 rule
    ('encounter', '123456',    'hl7_pid_required',        true, NULL, 'hl7', 'ADT'),
    -- Patient-level HL7 rule
    ('patient',   '123456789', 'hl7_patient_id_valid',    true, NULL, 'hl7', 'ADT')
ON CONFLICT (entity_type, entity_id, rule_type, source_format, event_type)
DO UPDATE SET
    compliance_flag = EXCLUDED.compliance_flag,
    reason_code     = EXCLUDED.reason_code,
    updated_at      = now();

-- ---------------------------------------------------------
-- FHIR Rules
-- ---------------------------------------------------------

INSERT INTO compliance_rules (
    entity_type, entity_id, rule_type, compliance_flag,
    reason_code, source_format, event_type
) VALUES
    -- Encounter-level FHIR rule
    ('encounter', '123456', 'fhir_encounter_status_required',   true, NULL, 'fhir', 'Encounter'),
    -- Patient reference rule
    ('patient',   '987654', 'fhir_patient_reference_valid',     true, NULL, 'fhir', 'Encounter')
ON CONFLICT (entity_type, entity_id, rule_type, source_format, event_type)
DO UPDATE SET
    compliance_flag = EXCLUDED.compliance_flag,
    reason_code     = EXCLUDED.reason_code,
    updated_at      = now();

-- ---------------------------------------------------------
-- Generic JSON Rules
-- ---------------------------------------------------------

INSERT INTO compliance_rules (
    entity_type, entity_id, rule_type, compliance_flag,
    reason_code, source_format, event_type
) VALUES
    -- Encounter-level generic rule
    ('encounter', '123456',   'generic_encounter_id_required', true, NULL, 'json', 'generic'),
    -- Member-level generic rule
    ('member',    'AET123987','generic_member_id_valid',       true, NULL, 'json', 'generic')
ON CONFLICT (entity_type, entity_id, rule_type, source_format, event_type)
DO UPDATE SET
    compliance_flag = EXCLUDED.compliance_flag,
    reason_code     = EXCLUDED.reason_code,
    updated_at      = now();
