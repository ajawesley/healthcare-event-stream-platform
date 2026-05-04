# tests/test_partitioner.py
from pyspark.sql import SparkSession
from partitioner import add_partition_columns


def test_add_partition_columns_with_metadata_event_date():
    spark = SparkSession.builder.master("local[1]").appName("test").getOrCreate()

    df = spark.createDataFrame(
        [
            {
                "event_id": "e1",
                "format": "fhir",
                "metadata": {"event_date": "2024-01-01"},
            }
        ]
    )

    out = add_partition_columns(df)
    cols = out.columns

    assert "event_date" in cols
    assert "format_partition" in cols

    row = out.collect()[0]
    assert str(row["format_partition"]) == "fhir"

    spark.stop()
