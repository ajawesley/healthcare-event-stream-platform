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
