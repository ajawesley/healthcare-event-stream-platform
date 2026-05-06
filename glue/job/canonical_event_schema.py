from pyspark.sql.types import (
    StructType,
    StructField,
    StringType
)

# ------------------------------------------------------------
# Canonical Event Schema (matches actual JSON)
# ------------------------------------------------------------

canonical_event_schema = StructType([
    StructField("canonical_event", StructType([
        StructField("event_id", StringType(), True),
        StructField("source_system", StringType(), True),
        StructField("format", StringType(), True),

        StructField("metadata", StructType([
            StructField("event_id", StringType(), True),
            StructField("source_system", StringType(), True),
        ]), True),

        StructField("patient", StructType([
            StructField("id", StringType(), True),
            StructField("first_name", StringType(), True),
            StructField("last_name", StringType(), True),
        ]), True),

        StructField("observation", StructType([
            StructField("code", StringType(), True),
        ]), True),

        StructField("raw_value", StringType(), True),
    ]), True),

    StructField("dispatched_at", StringType(), True),

    StructField("envelope", StructType([
        StructField("event_id", StringType(), True),
        StructField("event_type", StringType(), True),
        StructField("source_system", StringType(), True),
        StructField("produced_at", StringType(), True),
    ]), True),

    StructField("raw", StringType(), True),
])
