# tests/test_writer.py
from pyspark.sql import SparkSession
from writer import write_output
import os
import shutil


def test_write_output(tmp_path):
    spark = SparkSession.builder.master("local[1]").appName("test").getOrCreate()

    df = spark.createDataFrame(
        [
            {
                "event_id": "e1",
                "format": "fhir",
                "metadata": {},
                "event_date": "2024-01-01",
                "format_partition": "fhir",
            }
        ]
    )

    output_path = str(tmp_path / "out")
    write_output(df, output_path)

    assert os.path.exists(output_path)

    shutil.rmtree(output_path, ignore_errors=True)
    spark.stop()
