import pyspark.sql.functions as F
from pyspark.sql import SparkSession

from partitioner import add_partition_columns


def test_add_partition_columns_adds_expected_columns(spark: SparkSession):
    df = spark.createDataFrame(
        [("2024-01-01T12:00:00Z", "json")],
        ["produced_at", "format"],
    )

    result = add_partition_columns(df)

    cols = set(result.columns)
    assert "event_date" in cols
    assert "format_partition" in cols
    row = result.collect()[0]
    assert str(row["event_date"]) == "2024-01-01"
    assert row["format_partition"] == "json"
