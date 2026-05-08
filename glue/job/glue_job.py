# glue/job/glue_job.py
import sys
import time
from typing import Optional
from urllib.parse import urlparse

import boto3
from awsglue.utils import getResolvedOptions
from awsglue.context import GlueContext
from awsglue.job import Job
from pyspark.context import SparkContext
from pyspark.sql.functions import (
    col,
    lit,
    to_timestamp,
    input_file_name,
    to_json,
    udf,
)
from pyspark.sql.types import TimestampType

from partitioner import add_partition_columns
from writer import write_output
from error_writer import write_errors
from metrics import Metrics

from glue.job.hesp_logging import (
    get_logger,
    log_job_started,
    log_job_completed,
    log_job_failed,
)

log = get_logger()


def _get_arg_value(name: str) -> Optional[str]:
    flag = f"--{name}"
    if flag in sys.argv:
        idx = sys.argv.index(flag)
        if idx + 1 < len(sys.argv):
            return sys.argv[idx + 1]
    return None


# -------------------------------------------------------------------------
# Resolve arguments
# -------------------------------------------------------------------------
args = getResolvedOptions(
    sys.argv,
    ["JOB_NAME", "input_path", "output_base_path", "error_path"],
)

job_name = args["JOB_NAME"]
input_path = args["input_path"]
output_base_path = args["output_base_path"]
error_path = args["error_path"]

trace_id = _get_arg_value("trace_id")
span_id = _get_arg_value("span_id")
if trace_id or span_id:
    log.set_trace_context(trace_id, span_id)

# -------------------------------------------------------------------------
# Spark + Glue initialization
# -------------------------------------------------------------------------
sc = SparkContext()
glue_context = GlueContext(sc)
spark = glue_context.spark_session

job = Job(glue_context)
job.init(job_name, args)

metrics = Metrics()
start_time = time.time()

# -------------------------------------------------------------------------
# S3 client for metadata extraction
# -------------------------------------------------------------------------
s3 = boto3.client("s3")


@udf(TimestampType())
def get_s3_last_modified(s3_uri: str):
    if not s3_uri:
        return None
    parsed = urlparse(s3_uri)
    bucket = parsed.netloc
    key = parsed.path.lstrip("/")
    head = s3.head_object(Bucket=bucket, Key=key)
    return head["LastModified"]


# -------------------------------------------------------------------------
# Main job logic
# -------------------------------------------------------------------------
try:
    run_id = job.jobRunID
    log_job_started(job_name, run_id, trace_id=trace_id, span_id=span_id)

    log.info("reading_input_data", input_path=input_path)

    dyf_raw = glue_context.create_dynamic_frame.from_options(
        connection_type="s3",
        connection_options={"paths": [input_path]},
        format="json",
        format_options={"multiline": True},
    )

    total_records = dyf_raw.count()
    metrics.increment("total_records", total_records)
    log.info("records_loaded", total_records=total_records)

    if total_records == 0:
        log.info("no_records_found")
        metrics.log()
        job.commit()
        log_job_completed(job_name, run_id, duration_ms=0, trace_id=trace_id, span_id=span_id)
        sys.exit(0)

    df_raw = dyf_raw.toDF()

    # ---------------------------------------------------------------------
    # ⭐ Extract lineage + timestamps + trace context + S3 URI
    # ---------------------------------------------------------------------
    df = (
        df_raw
        .withColumn("trace_id", col("trace_id"))
        .withColumn("span_id", col("span_id"))
        .withColumn("lineage_json", to_json(col("lineage")))
        .withColumn("transmission_timestamp", to_timestamp(col("transmission_timestamp")))
        .withColumn("dispatched_at", to_timestamp(col("dispatched_at")))
        .withColumn("s3_uri", input_file_name())
    )

    # ---------------------------------------------------------------------
    # ⭐ Extract true durable write time from S3 metadata
    # ---------------------------------------------------------------------
    df = df.withColumn("s3_last_modified", get_s3_last_modified(col("s3_uri")))

    # ---------------------------------------------------------------------
    # Validation
    # ---------------------------------------------------------------------
    required_id_col = "canonical_event.event_id"
    required_format_col = "canonical_event.format"
    required_produced_at_col = "envelope.produced_at"

    valid_df = df.filter(
        col(required_id_col).isNotNull()
        & col(required_format_col).isNotNull()
        & col(required_produced_at_col).isNotNull()
    )

    invalid_df = df.filter(
        col(required_id_col).isNull()
        | col(required_format_col).isNull()
        | col(required_produced_at_col).isNull()
    ).withColumn(
        "error_reason",
        lit(
            "Missing required field: canonical_event.event_id, "
            "canonical_event.format, or envelope.produced_at"
        ),
    )

    valid_count = valid_df.count()
    invalid_count = invalid_df.count()

    metrics.increment("valid_records", valid_count)
    metrics.increment("invalid_records", invalid_count)

    log.info(
        "validation_results",
        valid_records=valid_count,
        invalid_records=invalid_count,
    )

    if invalid_count > 0:
        log.info("writing_invalid_records", error_path=error_path)
        write_errors(invalid_df, error_path)

    if valid_count == 0:
        log.info("no_valid_records")
        metrics.log()
        job.commit()
        log_job_completed(job_name, run_id, duration_ms=0, trace_id=trace_id, span_id=span_id)
        sys.exit(0)

    # ---------------------------------------------------------------------
    # ⭐ Add glue_processed_at for lineage dashboards
    # ---------------------------------------------------------------------
    glue_processed_at = time.time()
    valid_df = valid_df.withColumn("glue_processed_at", lit(glue_processed_at).cast(TimestampType()))

    # ---------------------------------------------------------------------
    # ⭐ Emit lineage completeness metrics
    # ---------------------------------------------------------------------
    metrics.increment("lineage_missing_trace_id", valid_df.filter(col("trace_id").isNull()).count())
    metrics.increment("lineage_missing_event_id", valid_df.filter(col(required_id_col).isNull()).count())
    metrics.increment("lineage_missing_transmission_ts", valid_df.filter(col("transmission_timestamp").isNull()).count())
    metrics.increment("lineage_missing_s3_last_modified", valid_df.filter(col("s3_last_modified").isNull()).count())

    # ---------------------------------------------------------------------
    # ⭐ Emit stage latency metrics (ms)
    # ---------------------------------------------------------------------
    def ms(col_a, col_b):
        return (col_b.cast("double") - col_a.cast("double")) * 1000.0

    latency_df = valid_df.select(
        ms(col("envelope.produced_at"), col("canonical_event.ingest_timestamp")).alias("ms_ingest_to_canonical"),
        ms(col("canonical_event.ingest_timestamp"), col("canonical_event.canonicalization_timestamp")).alias("ms_canonical_to_write"),
        ms(col("canonical_event.canonicalization_timestamp"), col("transmission_timestamp")).alias("ms_write_to_transmission"),
        ms(col("transmission_timestamp"), col("s3_last_modified")).alias("ms_transmission_to_s3"),
        ms(col("s3_last_modified"), col("glue_processed_at")).alias("ms_s3_to_glue"),
        ms(col("envelope.produced_at"), col("glue_processed_at")).alias("ms_end_to_end"),
    )

    for row in latency_df.collect():
        metrics.increment("pipeline_latency_ms_ingest_to_canonical", row["ms_ingest_to_canonical"])
        metrics.increment("pipeline_latency_ms_canonical_to_write", row["ms_canonical_to_write"])
        metrics.increment("pipeline_latency_ms_write_to_s3", row["ms_write_to_transmission"])
        metrics.increment("pipeline_latency_ms_s3_to_glue", row["ms_s3_to_glue"])
        metrics.increment("pipeline_latency_ms_end_to_end", row["ms_end_to_end"])

    # ---------------------------------------------------------------------
    # ⭐ Replay + Late Arrival Metrics
    # ---------------------------------------------------------------------
    replay_counts = (
        valid_df.groupBy("canonical_event.event_id")
        .count()
        .filter(col("count") > 1)
        .count()
    )
    metrics.increment("replay_events", replay_counts)

    late_arrivals = (
        valid_df.filter(
            (col("s3_last_modified").cast("long") - col("transmission_timestamp").cast("long")) > 300
        ).count()
    )
    metrics.increment("late_arrival_events", late_arrivals)

    # ---------------------------------------------------------------------
    # Flatten + partition
    # ---------------------------------------------------------------------
    log.info("flattening_fields_for_partitioning")

    flattened_df = (
        valid_df
        .withColumn("produced_at", col("envelope.produced_at"))
        .withColumn("format", col("canonical_event.format"))
        .withColumn("lineage_json", col("lineage_json"))
        .withColumn("transmission_timestamp", col("transmission_timestamp"))
        .withColumn("dispatched_at", col("dispatched_at"))
        .withColumn("s3_last_modified", col("s3_last_modified"))
        .withColumn("trace_id", col("trace_id"))
        .withColumn("span_id", col("span_id"))
        .withColumn("glue_processed_at", col("glue_processed_at"))
    )

    log.info("adding_partition_columns")
    partitioned_df = add_partition_columns(flattened_df)

    # ---------------------------------------------------------------------
    # Write output
    # ---------------------------------------------------------------------
    log.info("writing_valid_records", output_path=output_base_path)
    write_output(partitioned_df, output_base_path)

    metrics.log()
    job.commit()

    duration_ms = (time.time() - start_time) * 1000
    log_job_completed(job_name, run_id, duration_ms, trace_id=trace_id, span_id=span_id)

except Exception as e:
    log_job_failed(job_name, run_id, e, trace_id=trace_id, span_id=span_id)
    metrics.increment("job_failure", 1)
    metrics.log()
    raise
