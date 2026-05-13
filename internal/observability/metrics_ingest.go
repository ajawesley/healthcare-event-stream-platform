package observability

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	ingestMetricsOnce sync.Once

	ingestRequestsTotal metric.Int64Counter
	ingestErrorsTotal   metric.Int64Counter
	ingestLatencyMs     metric.Float64Histogram
	stageLatencyMs      metric.Float64Histogram
	s3PutLatencyMs      metric.Float64Histogram
)

func initIngestMetrics() {
	ingestMetricsOnce.Do(func() {
		meter = otel.Meter("hesp-ecs")

		ingestRequestsTotal, _ = meter.Int64Counter(
			"ingest_requests_total",
			metric.WithDescription("Total number of ingest requests received"),
		)

		ingestErrorsTotal, _ = meter.Int64Counter(
			"ingest_errors_total",
			metric.WithDescription("Total number of ingest requests that resulted in error"),
		)

		ingestLatencyMs, _ = meter.Float64Histogram(
			"ingest_request_latency_ms",
			metric.WithDescription("Latency of ingest requests in milliseconds"),
		)

		stageLatencyMs, _ = meter.Float64Histogram(
			"ingest_stage_latency_ms",
			metric.WithDescription("Latency of ingestion stages (normalize/transform/dispatch) in milliseconds"),
		)

		s3PutLatencyMs, _ = meter.Float64Histogram(
			"s3_put_latency_ms",
			metric.WithDescription("Latency of S3 PutObject calls in milliseconds"),
		)
	})
}

func IngestRequestStarted(ctx context.Context, route string) {
	initIngestMetrics()
	if ingestRequestsTotal == nil {
		return
	}
	ingestRequestsTotal.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("route", route),
		),
	)
}

func IngestRequestErrored(ctx context.Context, route, reason string) {
	initIngestMetrics()
	if ingestErrorsTotal == nil {
		return
	}
	ingestErrorsTotal.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("route", route),
			attribute.String("reason", reason),
		),
	)
}

func ObserveIngestLatency(ctx context.Context, route string, d time.Duration) {
	initIngestMetrics()
	if ingestLatencyMs == nil {
		return
	}
	ingestLatencyMs.Record(ctx, float64(d.Milliseconds()),
		metric.WithAttributes(
			attribute.String("route", route),
		),
	)
}

func ObserveStageLatency(ctx context.Context, stage, eventType, sourceSystem string, d time.Duration) {
	initIngestMetrics()
	if stageLatencyMs == nil {
		return
	}
	stageLatencyMs.Record(ctx, float64(d.Milliseconds()),
		metric.WithAttributes(
			attribute.String("stage", stage),
			attribute.String("event_type", eventType),
			attribute.String("source_system", sourceSystem),
		),
	)
}

func ObserveS3PutLatency(ctx context.Context, bucket, key string, d time.Duration, success bool) {
	initIngestMetrics()
	if s3PutLatencyMs == nil {
		return
	}
	s3PutLatencyMs.Record(ctx, float64(d.Milliseconds()),
		metric.WithAttributes(
			attribute.String("bucket", bucket),
			attribute.String("success", boolToString(success)),
		),
	)
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
