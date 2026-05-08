-- Detects schema drift by comparing the lineage schema_version
-- or by detecting unexpected fields in canonical_event.

CREATE OR REPLACE VIEW acmecorp_hesp_db.events_schema_drift_view AS
WITH base AS (
    SELECT
        event_id,
        trace_id,
        lineage_json,
        json_extract(lineage_json, '$.schema_version') AS schema_version,
        canonical_event,
        envelope,
        produced_at,
        s3_last_modified,
        glue_processed_at
    FROM acmecorp_hesp_db.events
)
SELECT
    event_id,
    trace_id,
    schema_version,
    s3_last_modified,
    glue_processed_at,
    -- Drift condition #1: schema_version changed
    CASE
        WHEN schema_version IS NOT NULL
             AND schema_version <> '1.0'
        THEN 'SCHEMA_VERSION_DRIFT'
        -- Drift condition #2: unexpected fields in canonical_event
        WHEN cardinality(map_keys(canonical_event)) > 20
        THEN 'CANONICAL_EVENT_FIELD_DRIFT'
        ELSE NULL
    END AS drift_type
FROM base
WHERE
    schema_version <> '1.0'
    OR cardinality(map_keys(canonical_event)) > 20
ORDER BY glue_processed_at DESC;
