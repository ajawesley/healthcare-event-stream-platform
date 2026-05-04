import sys
import logging
from awsglue.utils import getResolvedOptions
from awsglue.context import GlueContext
from awsglue.job import Job
from pyspark.context import SparkContext
from pyspark.sql.functions import col, lit

from canonical_event_schema import canonical_event_schema
from partitioner import add_partition_columns
from writer import write_output
from error_writer import write_errors
from metrics import Metrics

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# ------------------------------------------------------------
# Resolve arguments
# ------------------------------------------------------------
args = getResolvedOptions(
    sys.argv,
    ["JOB_NAME", "input_path", "output_base_path", "error_path"]
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
    # ------------------------------------------------------------
    # Load raw data
    # ------------------------------------------------------------
    logger.info(f"Reading input from: {input_path}")
    df_raw = spark.read.json(input_path)

    total_records = df_raw.count()
    metrics.increment("total_records", total_records)

    if total_records == 0:
        logger.info("No records found. Exiting gracefully.")
        metrics.log()
        job.commit()
        sys.exit(0)

    # ------------------------------------------------------------
    # Schema enforcement with try/except
    # ------------------------------------------------------------
    try:
        df = spark.createDataFrame(df_raw.rdd, schema=canonical_event_schema)
    except Exception as e:
        logger.error(f"Schema enforcement failed: {e}")
        df_raw = df_raw.withColumn("error_reason", lit(str(e)))
        write_errors(df_raw, error_path)
        metrics.increment("invalid_records", total_records)
        metrics.log()
        job.commit()
        sys.exit(0)

    # ------------------------------------------------------------
    # Valid vs invalid records
    # ------------------------------------------------------------
    valid_df = df.filter(
        col("event_id").isNotNull() &
        col("format").isNotNull() &
        col("produced_at").isNotNull()
    )

    invalid_df = df.filter(
        col("event_id").isNull() |
        col("format").isNull() |
        col("produced_at").isNull()
    ).withColumn(
        "error_reason",
        lit("Missing required field: event_id, format, or produced_at")
    )

    valid_count = valid_df.count()
    invalid_count = invalid_df.count()

    metrics.increment("valid_records", valid_count)
    metrics.increment("invalid_records", invalid_count)

    if invalid_count > 0:
        logger.info(f"Writing {invalid_count} invalid records to {error_path}")
        write_errors(invalid_df, error_path)

    if valid_count == 0:
        logger.info("No valid records to process. Exiting.")
        metrics.log()
        job.commit()
        sys.exit(0)

    # ------------------------------------------------------------
    # Partitioning
    # ------------------------------------------------------------
    logger.info("Adding partition columns")
    partitioned_df = add_partition_columns(valid_df)

    # ------------------------------------------------------------
    # Write output
    # ------------------------------------------------------------
    logger.info(f"Writing valid records to {output_base_path}")
    write_output(partitioned_df, output_base_path)

    metrics.log()
    job.commit()

except Exception as e:
    logger.error(f"Job failed: {e}")
    metrics.increment("job_failure", 1)
    metrics.log()
    raise
