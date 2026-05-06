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

//
// ────────────────────────────────────────────────────────────────────────────────
//   LAMBDA METRIC HELPERS
// ────────────────────────────────────────────────────────────────────────────────
//

// RecordLambdaInvocation increments the invocation counter.
func RecordLambdaInvocation() {
	if lambdaInvocationCount == nil {
		return
	}

	lambdaInvocationCount.Add(
		context.Background(),
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
		),
	)
}

// RecordLambdaError increments the error counter.
func RecordLambdaError() {
	if lambdaErrorCount == nil {
		return
	}

	lambdaErrorCount.Add(
		context.Background(),
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
		),
	)
}

// RecordLambdaLatency records Lambda invocation latency.
func RecordLambdaLatency(d time.Duration) {
	if lambdaLatency == nil {
		return
	}

	lambdaLatency.Record(
		context.Background(),
		float64(d.Milliseconds()),
		metric.WithAttributes(
			attrService,
			attrEnvironment,
		),
	)
}
