# tests/test_error_writer.py
from pyspark.sql import SparkSession
from error_writer import write_errors
import os
import shutil


def test_write_errors(tmp_path):
    spark = SparkSession.builder.master("local[1]").appName("test").getOrCreate()

    df = spark.createDataFrame(
        [
            {
                "event_id": "e1",
                "format": "fhir",
                "metadata": {},
            }
        ]
    )

    error_path = str(tmp_path / "errors")
    write_errors(df, error_path)

    assert os.path.exists(error_path)

    shutil.rmtree(error_path, ignore_errors=True)
    spark.stop()
