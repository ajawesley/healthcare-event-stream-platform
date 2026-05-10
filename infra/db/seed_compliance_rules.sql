CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS compliance_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    rule_type TEXT NOT NULL,
    compliance_flag BOOLEAN NOT NULL,
    reason_code TEXT,
    source_format TEXT NOT NULL,
    event_type TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

INSERT INTO compliance_rules (
    entity_type, entity_id, rule_type, compliance_flag,
    reason_code, source_format, event_type
) VALUES
    ('member', '123456789', 'x12_837_required_segments_present', true, NULL, 'x12', '837'),
    ('provider', '123456789', 'billing_provider_npi_valid', true, NULL, 'x12', '837'),
    ('member', '999999999', 'x12_837_required_segments_present', false, 'MISSING_NM1_IL', 'x12', '837');
