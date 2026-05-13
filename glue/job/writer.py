from pyspark.sql import DataFrame
from hesp_logging import get_logger

log = get_logger()

def write_output(df: DataFrame, output_base_path: str) -> None:
    """
    Writes valid canonical events to S3 in Parquet format,
    partitioned by:
      - event_date (YYYY-MM-DD)
      - format_partition (string)

    Adds pipeline-level metadata to Parquet footer:
      - service.name
      - pipeline.version
      - lineage.enabled
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

    # ----------------------------------------------------------------------
    # ⭐ Add Parquet file-level metadata (Phase 2 requirement)
    # ----------------------------------------------------------------------
    spark = df.sparkSession

    spark.conf.set("parquet.enable.summary-metadata", "true")
    spark.conf.set("parquet.metadata.pipeline", "hesp")
    spark.conf.set("parquet.metadata.pipeline_version", "1.0")
    spark.conf.set("parquet.metadata.lineage_enabled", "true")

    # ----------------------------------------------------------------------
    # Write Parquet output
    # ----------------------------------------------------------------------
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
