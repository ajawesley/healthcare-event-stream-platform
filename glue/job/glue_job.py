# glue/job/glue_job.py
import sys
import time
from typing import Optional

from awsglue.utils import getResolvedOptions
from awsglue.context import GlueContext
from awsglue.job import Job
from pyspark.context import SparkContext
from pyspark.sql.functions import col, lit

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

sc = SparkContext()
glue_context = GlueContext(sc)
spark = glue_context.spark_session

job = Job(glue_context)
job.init(job_name, args)

metrics = Metrics()
start_time = time.time()

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

    required_id_col = "canonical_event.event_id"
    required_format_col = "canonical_event.format"
    required_produced_at_col = "envelope.produced_at"

    valid_df = df_raw.filter(
        col(required_id_col).isNotNull()
        & col(required_format_col).isNotNull()
        & col(required_produced_at_col).isNotNull()
    )

    invalid_df = df_raw.filter(
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

    log.info("flattening_fields_for_partitioning")

    flattened_df = (
        valid_df.withColumn("produced_at", col("envelope.produced_at"))
        .withColumn("format", col("canonical_event.format"))
    )

    log.info("adding_partition_columns")
    partitioned_df = add_partition_columns(flattened_df)

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
