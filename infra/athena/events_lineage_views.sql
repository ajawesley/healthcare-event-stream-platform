-- events_lineage_views.sql
-- Assumes a Glue table like: acmecorp_hesp_db.events (adjust names as needed)

------------------------------------------------------------
-- 1. Base lineage view
------------------------------------------------------------
CREATE OR REPLACE VIEW acmecorp_hesp_db.events_lineage_view AS
SELECT
  event_id,
  trace_id,
  span_id,
  source_system,
  format,
  produced_at,
  ingest_timestamp,
  canonicalization_timestamp,
  write_timestamp,
  transmission_timestamp,
  dispatched_at,
  s3_last_modified,
  glue_processed_at,
  lineage_json
FROM acmecorp_hesp_db.events;

------------------------------------------------------------
-- 2. End-to-end latency view
------------------------------------------------------------
CREATE OR REPLACE VIEW acmecorp_hesp_db.events_latency_view AS
SELECT
  event_id,
  trace_id,
  produced_at,
  ingest_timestamp,
  canonicalization_timestamp,
  write_timestamp,
  transmission_timestamp,
  s3_last_modified,
  glue_processed_at,
  date_diff('millisecond', ingest_timestamp, canonicalization_timestamp)   AS ms_ingest_to_canonical,
  date_diff('millisecond', canonicalization_timestamp, write_timestamp)    AS ms_canonical_to_write,
  date_diff('millisecond', write_timestamp, transmission_timestamp)        AS ms_write_to_transmission,
  date_diff('millisecond', transmission_timestamp, s3_last_modified)       AS ms_transmission_to_s3,
  date_diff('millisecond', s3_last_modified, glue_processed_at)            AS ms_s3_to_glue,
  date_diff('millisecond', ingest_timestamp, glue_processed_at)            AS ms_end_to_end
FROM acmecorp_hesp_db.events;

------------------------------------------------------------
-- 3. Replay detection view
------------------------------------------------------------
CREATE OR REPLACE VIEW acmecorp_hesp_db.events_replay_view AS
WITH ranked AS (
  SELECT
    event_id,
    trace_id,
    transmission_timestamp,
    s3_last_modified,
    glue_processed_at,
    row_number() OVER (PARTITION BY event_id ORDER BY transmission_timestamp) AS rn,
    count(*)     OVER (PARTITION BY event_id) AS cnt
  FROM acmecorp_hesp_db.events
)
SELECT
  event_id,
  trace_id,
  transmission_timestamp,
  s3_last_modified,
  glue_processed_at,
  cnt AS replay_count
FROM ranked
WHERE cnt > 1
ORDER BY event_id, transmission_timestamp;

------------------------------------------------------------
-- 4. Late arrival detection view
------------------------------------------------------------
CREATE OR REPLACE VIEW acmecorp_hesp_db.events_late_arrival_view AS
SELECT
  event_id,
  trace_id,
  transmission_timestamp,
  s3_last_modified,
  glue_processed_at,
  date_diff('second', transmission_timestamp, s3_last_modified) AS delay_seconds
FROM acmecorp_hesp_db.events
WHERE date_diff('minute', transmission_timestamp, s3_last_modified) > 5
ORDER BY delay_seconds DESC;

------------------------------------------------------------
-- 5. Schema drift detection (simple version)
------------------------------------------------------------
CREATE OR REPLACE VIEW acmecorp_hesp_db.events_schema_drift_view AS
SELECT
  event_id,
  trace_id,
  lineage_json,
  json_extract(lineage_json, '$.schema_version') AS schema_version
FROM acmecorp_hesp_db.events
WHERE json_extract(lineage_json, '$.schema_version') IS NOT NULL;
