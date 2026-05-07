package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter metric.Meter

	// ECS HTTP metrics
	httpRequestCount   metric.Int64Counter
	httpRequestLatency metric.Float64Histogram
	httpErrorCount     metric.Int64Counter

	// External dependency metrics
	dbQueryLatency    metric.Float64Histogram
	redisLatency      metric.Float64Histogram
	s3Latency         metric.Float64Histogram
	httpClientLatency metric.Float64Histogram

	// Lambda metrics
	lambdaInvocationCount metric.Int64Counter
	lambdaErrorCount      metric.Int64Counter
	lambdaLatency         metric.Float64Histogram

	// Resource attributes
	attrService     attribute.KeyValue
	attrEnvironment attribute.KeyValue
)

// InitMetrics initializes OTEL metrics for ECS + Lambda.
func InitMetrics(serviceName, environment string) {
	meter = otel.Meter("hesp")

	attrService = attribute.String("service.name", serviceName)
	attrEnvironment = attribute.String("deployment.environment", environment)

	// -------------------------------
	// ECS HTTP metrics
	// -------------------------------
	httpRequestCount, _ = meter.Int64Counter(
		"hesp_http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)

	httpRequestLatency, _ = meter.Float64Histogram(
		"hesp_http_request_latency_ms",
		metric.WithDescription("HTTP request latency in milliseconds"),
	)

	httpErrorCount, _ = meter.Int64Counter(
		"hesp_http_errors_total",
		metric.WithDescription("Total number of HTTP 5xx errors"),
	)

	// -------------------------------
	// External dependency metrics
	// -------------------------------
	dbQueryLatency, _ = meter.Float64Histogram(
		"db_query_latency_ms",
		metric.WithDescription("Latency of database queries in milliseconds"),
	)

	redisLatency, _ = meter.Float64Histogram(
		"redis_latency_ms",
		metric.WithDescription("Latency of Redis operations in milliseconds"),
	)

	s3Latency, _ = meter.Float64Histogram(
		"s3_operation_latency_ms",
		metric.WithDescription("Latency of S3 operations in milliseconds"),
	)

	httpClientLatency, _ = meter.Float64Histogram(
		"http_client_latency_ms",
		metric.WithDescription("Latency of outbound HTTP calls in milliseconds"),
	)

	// -------------------------------
	// Lambda metrics
	// -------------------------------
	lambdaInvocationCount, _ = meter.Int64Counter(
		"hesp_lambda_invocations_total",
		metric.WithDescription("Total number of Lambda invocations"),
	)

	lambdaErrorCount, _ = meter.Int64Counter(
		"hesp_lambda_errors_total",
		metric.WithDescription("Total number of Lambda errors"),
	)

	lambdaLatency, _ = meter.Float64Histogram(
		"hesp_lambda_latency_ms",
		metric.WithDescription("Lambda invocation latency in milliseconds"),
	)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//   ECS METRIC HELPERS
// ────────────────────────────────────────────────────────────────────────────────
//

// RecordRequestCount increments the HTTP request counter.
func RecordRequestCount(method string, status int) {
	if httpRequestCount == nil {
		return
	}

	httpRequestCount.Add(
		context.Background(),
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("http.method", method),
			attribute.Int("http.status_code", status),
		),
	)
}

// RecordRequestLatency records HTTP request latency.
func RecordRequestLatency(method string, status int, d time.Duration) {
	if httpRequestLatency == nil {
		return
	}

	httpRequestLatency.Record(
		context.Background(),
		float64(d.Milliseconds()),
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("http.method", method),
			attribute.Int("http.status_code", status),
		),
	)
}

// RecordHttpError increments the HTTP error counter.
func RecordHttpError(method string, status int) {
	if httpErrorCount == nil {
		return
	}

	httpErrorCount.Add(
		context.Background(),
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("http.method", method),
			attribute.Int("http.status_code", status),
		),
	)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//   EXTERNAL DEPENDENCY HELPERS
// ────────────────────────────────────────────────────────────────────────────────
//

func RecordDBLatency(op, table string, d time.Duration) {
	if dbQueryLatency == nil {
		return
	}

	dbQueryLatency.Record(
		context.Background(),
		float64(d.Milliseconds()),
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("db.operation", op),
			attribute.String("db.table", table),
		),
	)
}

func RecordRedisLatency(cmd string, d time.Duration) {
	if redisLatency == nil {
		return
	}

	redisLatency.Record(
		context.Background(),
		float64(d.Milliseconds()),
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("redis.command", cmd),
		),
	)
}

func RecordS3Latency(op, bucket string, d time.Duration) {
	if s3Latency == nil {
		return
	}

	s3Latency.Record(
		context.Background(),
		float64(d.Milliseconds()),
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("s3.operation", op),
			attribute.String("s3.bucket", bucket),
		),
	)
}

func RecordHttpClientLatency(method, host string, status int, d time.Duration) {
	if httpClientLatency == nil {
		return
	}

	httpClientLatency.Record(
		context.Background(),
		float64(d.Milliseconds()),
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("http.method", method),
			attribute.String("http.host", host),
			attribute.Int("http.status_code", status),
		),
	)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//   LAMBDA METRIC HELPERS
// ────────────────────────────────────────────────────────────────────────────────
//

func RecordLambdaInvocation() {
	if lambdaInvocationCount == nil {
		return
	}

	lambdaInvocationCount.Add(
		context.Background(),
		1,
		metric.WithAttributes(attrService, attrEnvironment),
	)
}

func RecordLambdaError() {
	if lambdaErrorCount == nil {
		return
	}

	lambdaErrorCount.Add(
		context.Background(),
		1,
		metric.WithAttributes(attrService, attrEnvironment),
	)
}

func RecordLambdaLatency(d time.Duration) {
	if lambdaLatency == nil {
		return
	}

	lambdaLatency.Record(
		context.Background(),
		float64(d.Milliseconds()),
		metric.WithAttributes(attrService, attrEnvironment),
	)
}
