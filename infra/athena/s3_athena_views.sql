-- s3_athena_views.sql
-- Assumes you have an external table over CloudTrail S3 data events.

CREATE OR REPLACE VIEW acmecorp_hesp_db.s3_ingestion_events AS
SELECT
  eventtime,
  useridentity.principalid          AS principal_id,
  requestparameters.bucketName      AS bucket_name,
  requestparameters.key             AS object_key,
  additionalEventData.bytesTransferredIn AS size_bytes,
  eventname                         AS event_name,
  sourceipaddress                   AS source_ip
FROM cloudtrail_s3_events
WHERE eventname IN ('PutObject', 'CompleteMultipartUpload');

CREATE OR REPLACE VIEW acmecorp_hesp_db.s3_ingestion_hourly AS
SELECT
  date_trunc('hour', eventtime) AS hour,
  bucket_name,
  count(*)                      AS object_count,
  sum(size_bytes)               AS total_bytes
FROM acmecorp_hesp_db.s3_ingestion_events
GROUP BY 1, 2
ORDER BY hour DESC;
