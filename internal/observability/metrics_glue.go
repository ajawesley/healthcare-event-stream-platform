package observability

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	glueMetricsOnce sync.Once

	glueRecordsTotal   metric.Int64Counter
	glueRecordsValid   metric.Int64Counter
	glueRecordsInvalid metric.Int64Counter
	glueParquetWriteMs metric.Float64Histogram
	glueS3ReadBytes    metric.Float64Histogram
	glueS3WriteBytes   metric.Float64Histogram
)

func InitGlueMetrics() {
	glueMetricsOnce.Do(func() {
		meter = otel.Meter("hesp-glue")

		glueRecordsTotal, _ = meter.Int64Counter(
			"glue_records_total",
			metric.WithDescription("Total records processed by Glue ingestion"),
		)

		glueRecordsValid, _ = meter.Int64Counter(
			"glue_records_valid_total",
			metric.WithDescription("Valid records processed by Glue ingestion"),
		)

		glueRecordsInvalid, _ = meter.Int64Counter(
			"glue_records_invalid_total",
			metric.WithDescription("Invalid records processed by Glue ingestion"),
		)

		glueParquetWriteMs, _ = meter.Float64Histogram(
			"glue_parquet_write_ms",
			metric.WithDescription("Time taken to write Parquet files"),
		)

		glueS3ReadBytes, _ = meter.Float64Histogram(
			"glue_s3_read_bytes",
			metric.WithDescription("S3 read throughput in bytes"),
		)

		glueS3WriteBytes, _ = meter.Float64Histogram(
			"glue_s3_write_bytes",
			metric.WithDescription("S3 write throughput in bytes"),
		)
	})
}

func GlueRecordTotal(n int64) {
	glueRecordsTotal.Add(context.Background(), n)
}

func GlueRecordValid(n int64) {
	glueRecordsValid.Add(context.Background(), n)
}

func GlueRecordInvalid(n int64) {
	glueRecordsInvalid.Add(context.Background(), n)
}

func GlueParquetWriteDuration(d time.Duration) {
	glueParquetWriteMs.Record(context.Background(), float64(d.Milliseconds()))
}

func GlueS3Read(sizeBytes int64) {
	glueS3ReadBytes.Record(context.Background(), float64(sizeBytes))
}

func GlueS3Write(sizeBytes int64) {
	glueS3WriteBytes.Record(context.Background(), float64(sizeBytes))
}
