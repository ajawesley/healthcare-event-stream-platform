from pyspark.sql import DataFrame
from pyspark.sql.functions import col, to_timestamp, to_date

from hesp_logging import get_logger

log = get_logger()

def add_partition_columns(df: DataFrame) -> DataFrame:
    """
    Adds partition columns required by the ingestion pipeline.
    """

    log.info("partitioner_start")

    df_with_ts = df.withColumn(
        "produced_at_ts",
        to_timestamp(col("produced_at"))
    )
    log.info("partitioner_parsed_timestamp")

    df_with_date = df_with_ts.withColumn(
        "event_date",
        to_date(col("produced_at_ts"))
    )
    log.info("partitioner_added_event_date")

    df_final = df_with_date.withColumn(
        "format_partition",
        col("format")
    )
    log.info("partitioner_added_format_partition")

    log.info("partitioner_complete")

    return df_final.drop("produced_at_ts")
