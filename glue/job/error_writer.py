import logging
from pyspark.sql import DataFrame
from pyspark.sql.functions import current_date

logger = logging.getLogger(__name__)

def write_errors(df: DataFrame, error_path: str) -> None:
    """
    Writes invalid or malformed records to S3 in JSON format.

    Adds:
      - error_date (partition key)
      - error_reason (must be added upstream)

    Does NOT enforce schema — error records may contain arbitrary fields.
    """

    if "error_reason" not in df.columns:
        raise ValueError(
            "error_reason column missing. "
            "Ensure glue_job.py attaches an error_reason before calling write_errors()."
        )

    df_with_date = df.withColumn("error_date", current_date())

    logger.info(f"Writing {df_with_date.count()} error records to {error_path}")

    (
        df_with_date.write
        .mode("append")
        .partitionBy("error_date")
        .json(error_path)
    )

    logger.info("Error write completed successfully")
