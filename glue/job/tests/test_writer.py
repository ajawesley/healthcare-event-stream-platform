import os
from pathlib import Path

from writer import write_output


def test_write_output_writes_parquet(tmp_path, spark):
    df = spark.createDataFrame(
        [("2024-01-01", "json", 1)],
        ["event_date", "format_partition", "value"],
    )

    output = tmp_path / "out"
    write_output(df, str(output))

    files = list(Path(output).rglob("*.parquet"))
    assert files
