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
