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
    expr,
)
from pyspark.sql.types import TimestampType

from partitioner import add_partition_columns
from writer import write_output
from error_writer import write_errors
from metrics import Metrics

from hesp_logging import (
    get_logger,
    log_job_started,
    log_job_completed,
    log_job_failed,
)

# -------------------------------------------------------------------------
# No‑Op Tracer (replaces OpenTelemetry)
# -------------------------------------------------------------------------
from contextlib import nullcontext

class NoopTracer:
    def start_as_current_span(self, name):
        return nullcontext()

tracer = NoopTracer()

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

job_run_id = _get_arg_value("jobRunId") or f"manual-{int(time.time())}"
run_id = job_run_id

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
# SAFE UDF — boto3 client created inside worker
# -------------------------------------------------------------------------
@udf(TimestampType())
def get_s3_last_modified(s3_uri: str):
    if not s3_uri:
        return None

    import boto3
    from urllib.parse import urlparse

    s3 = boto3.client("s3")

    parsed = urlparse(s3_uri)
    bucket = parsed.netloc
    key = parsed.path.lstrip("/")

    head = s3.head_object(Bucket=bucket, Key=key)
    return head["LastModified"]


# -------------------------------------------------------------------------
# MAIN JOB LOGIC
# -------------------------------------------------------------------------
try:
    print("🔥🔥🔥 GLUE SCRIPT EXECUTED 🔥🔥🔥")
    print(f"job_name: {job_name}")
    print(f"input_path: {input_path}")
    print(f"output_base_path: {output_base_path}")
    print(f"error_path: {error_path}")
    print(f"run_id: {run_id}")
    print(f"trace_id: {trace_id}")
    print(f"span_id: {span_id}")

    with tracer.start_as_current_span("glue_job_root"):
        log_job_started(job_name, run_id, trace_id=trace_id, span_id=span_id)

        # -------------------------------
        # Read input
        # -------------------------------
        dyf_raw = glue_context.create_dynamic_frame.from_options(
            connection_type="s3",
            connection_options={"paths": [input_path]},
            format="json",
            format_options={"multiline": True},
        )

        total_records = dyf_raw.count()
        metrics.increment("total_records", total_records)

        if total_records == 0:
            metrics.log()
            job.commit()
            log_job_completed(job_name, run_id, duration_ms=0)
            sys.exit(0)

        df_raw = dyf_raw.toDF()

        # -------------------------------
        # Extract lineage timestamps from array
        # -------------------------------
        df = (
            df_raw
            .withColumn("trace_id", col("trace_id"))
            .withColumn("span_id", col("span_id"))
            .withColumn("lineage_json", to_json(col("lineage")))
            .withColumn("transmission_timestamp", to_timestamp(col("transmission_timestamp")))
            .withColumn("dispatched_at", to_timestamp(col("dispatched_at")))
            .withColumn("s3_uri", input_file_name())
            # Extract lineage timestamps
            .withColumn("lineage_ingest_ts",
                        expr("filter(lineage, x -> x.name = 'ingest')[0].timestamp"))
            .withColumn("lineage_normalized_ts",
                        expr("filter(lineage, x -> x.name = 'normalized')[0].timestamp"))
            .withColumn("lineage_canonicalized_ts",
                        expr("filter(lineage, x -> x.name = 'canonicalized')[0].timestamp"))
            .withColumn("lineage_ingest_ts", to_timestamp("lineage_ingest_ts"))
            .withColumn("lineage_normalized_ts", to_timestamp("lineage_normalized_ts"))
            .withColumn("lineage_canonicalized_ts", to_timestamp("lineage_canonicalized_ts"))
        )

        # -------------------------------
        # Extract S3 durable write time
        # -------------------------------
        df = df.withColumn("s3_last_modified", get_s3_last_modified(col("s3_uri")))

        # -------------------------------
        # Validation
        # -------------------------------
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
            lit("Missing required canonical_event or envelope fields"),
        )

        if invalid_df.count() > 0:
            write_errors(invalid_df, error_path)

        if valid_df.count() == 0:
            metrics.log()
            job.commit()
            log_job_completed(job_name, run_id, duration_ms=0)
            sys.exit(0)

        # -------------------------------
        # Add glue_processed_at
        # -------------------------------
        glue_processed_at = time.time()
        valid_df = valid_df.withColumn(
            "glue_processed_at",
            lit(glue_processed_at).cast(TimestampType())
        )

        # -------------------------------
        # Latency metrics (corrected)
        # -------------------------------
        def ms(col_a, col_b):
            return (col_b.cast("double") - col_a.cast("double")) * 1000.0

        latency_df = valid_df.select(
            ms(col("envelope.produced_at"), col("lineage_ingest_ts")).alias("ms_ingest_to_normalized"),
            ms(col("lineage_ingest_ts"), col("lineage_canonicalized_ts")).alias("ms_normalized_to_canonicalized"),
            ms(col("lineage_canonicalized_ts"), col("transmission_timestamp")).alias("ms_canonicalized_to_transmission"),
            ms(col("transmission_timestamp"), col("s3_last_modified")).alias("ms_transmission_to_s3"),
            ms(col("s3_last_modified"), col("glue_processed_at")).alias("ms_s3_to_glue"),
            ms(col("envelope.produced_at"), col("glue_processed_at")).alias("ms_end_to_end"),
        )

        for row in latency_df.collect():
            for k, v in row.asDict().items():
                if v is not None:
                    metrics.increment(k, float(v))


        # -------------------------------
        # Flatten + partition
        # -------------------------------
        flattened_df = (
            valid_df
            .withColumn("produced_at", col("envelope.produced_at"))
            .withColumn("format", col("canonical_event.format"))
        )

        partitioned_df = add_partition_columns(flattened_df)

        # -------------------------------
        # Write output
        # -------------------------------
        write_output(partitioned_df, output_base_path)

        metrics.log()
        job.commit()

        duration_ms = (time.time() - start_time) * 1000
        log_job_completed(job_name, run_id, duration_ms)

except Exception as e:
    log_job_failed(job_name, run_id, e)
    metrics.increment("job_failure", 1)
    metrics.log()
    raise
