from pyspark.sql.types import (
    StructType,
    StructField,
    StringType,
    TimestampType,
    MapType,
    StructType,
    StructField
)

# ------------------------------------------------------------
# Canonical Event Schema
# ------------------------------------------------------------

canonical_event_schema = StructType([
    StructField("event_id", StringType(), False),
    StructField("format", StringType(), False),

    # NEW: ISO-8601 timestamp from ingestion service
    StructField("produced_at", TimestampType(), False),

    StructField("patient", StructType([
        StructField("id", StringType(), True),
        StructField("name", StringType(), True),
        StructField("dob", StringType(), True),
        StructField("gender", StringType(), True)
    ]), True),

    StructField("observation", StructType([
        StructField("type", StringType(), True),
        StructField("value", StringType(), True),
        StructField("unit", StringType(), True)
    ]), True),

    StructField("metadata", MapType(StringType(), StringType()), True)
])
