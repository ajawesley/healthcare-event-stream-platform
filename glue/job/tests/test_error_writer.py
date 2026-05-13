from pyspark.sql import SparkSession
from glue.job.error_writer import write_errors
import os
import shutil


def test_write_errors(tmp_path):
    spark = (
        SparkSession.builder.master("local[1]")
        .appName("test_error_writer")
        .getOrCreate()
    )

    # Must include error_reason (required by updated writer)
    df = spark.createDataFrame(
        [
            {
                "event_id": "e1",
                "format": "fhir",
                "metadata": "{}",
                "error_reason": "missing required fields",
            }
        ]
    )

    error_path = str(tmp_path / "errors")
    write_errors(df, error_path)

    # Validate output directory exists
    assert os.path.exists(error_path)

    shutil.rmtree(error_path, ignore_errors=True)
    spark.stop()
