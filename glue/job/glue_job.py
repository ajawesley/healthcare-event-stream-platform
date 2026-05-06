import sys
import logging

from awsglue.utils import getResolvedOptions
from awsglue.context import GlueContext
from awsglue.job import Job
from awsglue.dynamicframe import DynamicFrame
from pyspark.context import SparkContext
from pyspark.sql.functions import col, lit

from partitioner import add_partition_columns
from writer import write_output
from error_writer import write_errors
from metrics import Metrics

# ------------------------------------------------------------
# Logging
# ------------------------------------------------------------
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# ------------------------------------------------------------
# Resolve arguments
# ------------------------------------------------------------
args = getResolvedOptions(
    sys.argv,
    ["JOB_NAME", "input_path", "output_base_path", "error_path"],
)

input_path = args["input_path"]
output_base_path = args["output_base_path"]
error_path = args["error_path"]

# ------------------------------------------------------------
# Spark / Glue setup
# ------------------------------------------------------------
sc = SparkContext()
glue_context = GlueContext(sc)
spark = glue_context.spark_session

job = Job(glue_context)
job.init(args["JOB_NAME"], args)

metrics = Metrics()

try:
    # --------------------------------------------------------
    # Load raw data via DynamicFrame (self-describing)
    # --------------------------------------------------------
    logger.info(f"Reading input from: {input_path}")

    dyf_raw = glue_context.create_dynamic_frame.from_options(
        connection_type="s3",
        connection_options={"paths": [input_path]},
        format="json",
        format_options={"multiline": True}
    )

    total_records = dyf_raw.count()
    metrics.increment("total_records", total_records)
    logger.info(f"Total records read: {total_records}")

    if total_records == 0:
        logger.info("No records found. Exiting gracefully.")
        metrics.log()
        job.commit()
        sys.exit(0)

    # Convert to DataFrame for validation / partitioning
    df_raw = dyf_raw.toDF()

    # --------------------------------------------------------
    # Valid vs invalid records (NESTED SCHEMA, BUT FLEXIBLE)
    # --------------------------------------------------------
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
        )
    )

    valid_count = valid_df.count()
    invalid_count = invalid_df.count()

    metrics.increment("valid_records", valid_count)
    metrics.increment("invalid_records", invalid_count)

    logger.info(f"Valid records: {valid_count}")
    logger.info(f"Invalid records: {invalid_count}")

    if invalid_count > 0:
        logger.info(f"Writing {invalid_count} invalid records to {error_path}")
        write_errors(invalid_df, error_path)

    if valid_count == 0:
        logger.info("No valid records to process. Exiting.")
        metrics.log()
        job.commit()
        sys.exit(0)

    # --------------------------------------------------------
    # Flatten nested fields for partitioner
    #   - produced_at  <- envelope.produced_at
    #   - format       <- canonical_event.format
    # --------------------------------------------------------
    logger.info("Flattening nested fields for partitioning.")
    flattened_df = (
        valid_df
        .withColumn("produced_at", col("envelope.produced_at"))
        .withColumn("format", col("canonical_event.format"))
    )

    # --------------------------------------------------------
    # Partitioning
    # --------------------------------------------------------
    logger.info("Adding partition columns.")
    partitioned_df = add_partition_columns(flattened_df)

    # --------------------------------------------------------
    # Write output
    # --------------------------------------------------------
    logger.info(f"Writing valid records to {output_base_path}")
    write_output(partitioned_df, output_base_path)

    metrics.log()
    job.commit()
    logger.info("Job completed successfully.")

except Exception as e:
    logger.error(f"Job failed: {e}", exc_info=True)
    metrics.increment("job_failure", 1)
    metrics.log()
    raise
