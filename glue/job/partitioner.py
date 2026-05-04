from pyspark.sql import DataFrame
from pyspark.sql.functions import (
    col,
    to_timestamp,
    to_date,
    lit
)

def add_partition_columns(df: DataFrame) -> DataFrame:
    """
    Adds partition columns required by the ingestion pipeline:
      - event_date: derived from produced_at (ISO-8601 timestamp)
      - format_partition: mirrors the `format` field

    Assumes:
      - produced_at is a valid ISO-8601 string
      - format is non-null (validated earlier in glue_job.py)
    """

    # Parse ISO-8601 timestamp into Spark timestamp
    df_with_ts = df.withColumn(
        "produced_at_ts",
        to_timestamp(col("produced_at"))
    )

    # Derive event_date (YYYY-MM-DD)
    df_with_date = df_with_ts.withColumn(
        "event_date",
        to_date(col("produced_at_ts"))
    )

    # Derive format_partition
    df_final = df_with_date.withColumn(
        "format_partition",
        col("format")
    )

    return df_final.drop("produced_at_ts")
