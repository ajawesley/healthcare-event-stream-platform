import pytest
from pyspark.sql import SparkSession

from error_writer import write_errors


def test_write_errors_requires_error_reason(spark: SparkSession, tmp_path):
    df = spark.createDataFrame([(1,)], ["value"])
    with pytest.raises(ValueError):
        write_errors(df, str(tmp_path / "errors"))
