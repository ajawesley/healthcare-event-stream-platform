from pyspark.sql import DataFrame
from glue.job.hesp_logging import get_logger

log = get_logger()

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

    count = df.count()

    log.info(
        "writing_output_records",
        output_path=output_base_path,
        record_count=count,
    )

    (
        df.write
        .mode("append")
        .partitionBy("event_date", "format_partition")
        .parquet(output_base_path)
    )

    log.info(
        "output_write_completed",
        output_path=output_base_path,
        record_count=count,
    )
