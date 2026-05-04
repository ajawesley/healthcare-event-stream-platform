import logging
from pyspark.sql import DataFrame

logger = logging.getLogger(__name__)

def write_output(df: DataFrame, output_base_path: str) -> None:
    """
    Writes valid canonical events to S3 in Parquet format,
    partitioned by:
      - event_date (YYYY-MM-DD)
      - format_partition (string)

    Assumes these columns were added by partitioner.py.
    """

    required_cols = {"event_date", "format_partition"}
    missing = required_cols - set(df.columns)

    if missing:
        raise ValueError(
            f"Missing required partition columns: {missing}. "
            "Ensure add_partition_columns() was applied."
        )

    logger.info(f"Writing output to {output_base_path}")

    (
        df.write
        .mode("append")
        .partitionBy("event_date", "format_partition")
        .parquet(output_base_path)
    )

    logger.info("Output write completed successfully")
