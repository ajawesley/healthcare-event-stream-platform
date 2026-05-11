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

	// Lineage metrics
	lineageLatency       metric.Float64Histogram
	lineageEvents        metric.Int64Counter
	lineageFailures      metric.Int64Counter
	lineageExportErrors  metric.Int64Counter
	lineageReadinessPing metric.Int64Counter

	// Compliance metrics
	complianceRuleHits        metric.Int64Counter
	complianceRuleMisses      metric.Int64Counter
	complianceFallbacks       metric.Int64Counter
	complianceErrors          metric.Int64Counter
	complianceLookupLatency   metric.Float64Histogram
	complianceBackendFailures metric.Int64Counter

	// Resilience metrics
	retryAttempts          metric.Int64Counter
	timeoutTotal           metric.Int64Counter
	circuitBreakerOpen     metric.Int64Counter
	circuitBreakerHalfOpen metric.Int64Counter
	circuitBreakerClosed   metric.Int64Counter
	backpressureRejections metric.Int64Counter

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

	// -------------------------------
	// Lineage metrics
	// -------------------------------
	lineageLatency, _ = meter.Float64Histogram(
		"hesp_lineage_stage_latency_ms",
		metric.WithDescription("Latency per lineage stage in milliseconds"),
	)

	lineageEvents, _ = meter.Int64Counter(
		"hesp_lineage_events_total",
		metric.WithDescription("Total number of lineage events observed"),
	)

	lineageFailures, _ = meter.Int64Counter(
		"hesp_lineage_failures_total",
		metric.WithDescription("Total number of lineage failures"),
	)

	lineageExportErrors, _ = meter.Int64Counter(
		"hesp_lineage_export_errors_total",
		metric.WithDescription("Total number of lineage export errors"),
	)

	lineageReadinessPing, _ = meter.Int64Counter(
		"hesp_lineage_readiness_pings_total",
		metric.WithDescription("Total number of lineage readiness probe pings"),
	)

	// -------------------------------
	// Compliance metrics
	// -------------------------------
	complianceRuleHits, _ = meter.Int64Counter(
		"hesp_compliance_rule_hits_total",
		metric.WithDescription("Total number of compliance rule hits"),
	)

	complianceRuleMisses, _ = meter.Int64Counter(
		"hesp_compliance_rule_misses_total",
		metric.WithDescription("Total number of compliance rule misses (no rule found)"),
	)

	complianceFallbacks, _ = meter.Int64Counter(
		"hesp_compliance_fallback_total",
		metric.WithDescription("Total number of compliance fallbacks applied"),
	)

	complianceErrors, _ = meter.Int64Counter(
		"hesp_compliance_errors_total",
		metric.WithDescription("Total number of compliance errors"),
	)

	complianceLookupLatency, _ = meter.Float64Histogram(
		"hesp_compliance_lookup_latency_ms",
		metric.WithDescription("Latency of compliance rule lookups in milliseconds"),
	)

	complianceBackendFailures, _ = meter.Int64Counter(
		"hesp_compliance_backend_failures_total",
		metric.WithDescription("Total number of compliance backend failures (redis/postgres/dynamodb)"),
	)

	// -------------------------------
	// Resilience metrics
	// -------------------------------
	retryAttempts, _ = meter.Int64Counter(
		"hesp_resilience_retry_attempts_total",
		metric.WithDescription("Total number of retry attempts"),
	)

	timeoutTotal, _ = meter.Int64Counter(
		"hesp_resilience_timeouts_total",
		metric.WithDescription("Total number of timeouts"),
	)

	circuitBreakerOpen, _ = meter.Int64Counter(
		"hesp_resilience_circuit_breaker_open_total",
		metric.WithDescription("Total number of times a circuit breaker opened"),
	)

	circuitBreakerHalfOpen, _ = meter.Int64Counter(
		"hesp_resilience_circuit_breaker_half_open_total",
		metric.WithDescription("Total number of times a circuit breaker entered half-open state"),
	)

	circuitBreakerClosed, _ = meter.Int64Counter(
		"hesp_resilience_circuit_breaker_closed_total",
		metric.WithDescription("Total number of times a circuit breaker closed"),
	)

	backpressureRejections, _ = meter.Int64Counter(
		"hesp_resilience_backpressure_rejections_total",
		metric.WithDescription("Total number of requests rejected due to backpressure"),
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

// ObserveDependencyLatency records latency for any external dependency.
// Example dependency types: "db", "redis", "s3", "http", "kafka", etc.
func ObserveDependencyLatency(ctx context.Context, depType, operation, target string, start time.Time) {
	if httpClientLatency == nil {
		return
	}

	d := time.Since(start)

	httpClientLatency.Record(ctx, float64(d.Milliseconds()),
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("dependency.type", depType),
			attribute.String("dependency.operation", operation),
			attribute.String("dependency.target", target),
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

//
// ────────────────────────────────────────────────────────────────────────────────
//   LINEAGE METRIC HELPERS
// ────────────────────────────────────────────────────────────────────────────────
//

// ObserveLineageLatency records latency for a lineage stage.
func ObserveLineageLatency(ctx context.Context, stage string, start time.Time) {
	if lineageLatency == nil {
		return
	}

	d := time.Since(start)

	lineageLatency.Record(
		ctx,
		float64(d.Milliseconds()),
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("lineage.stage", stage),
		),
	)
}

// IncrementLineageEvent increments lineage event counter.
func IncrementLineageEvent(ctx context.Context, stage, sourceSystem, eventType string) {
	if lineageEvents == nil {
		return
	}

	lineageEvents.Add(
		ctx,
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("lineage.stage", stage),
			attribute.String("source_system", sourceSystem),
			attribute.String("event_type", eventType),
		),
	)
}

// IncrementLineageFailure increments lineage failure counter.
func IncrementLineageFailure(ctx context.Context, stage, reason string) {
	if lineageFailures == nil {
		return
	}

	lineageFailures.Add(
		ctx,
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("lineage.stage", stage),
			attribute.String("failure_reason", reason),
		),
	)
}

// IncrementLineageExportError increments lineage export error counter.
func IncrementLineageExportError(ctx context.Context, backend string) {
	if lineageExportErrors == nil {
		return
	}

	lineageExportErrors.Add(
		ctx,
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("backend", backend),
		),
	)
}

// RecordLineageReadinessPing increments lineage readiness probe counter.
func RecordLineageReadinessPing() {
	if lineageReadinessPing == nil {
		return
	}

	lineageReadinessPing.Add(
		context.Background(),
		1,
		metric.WithAttributes(attrService, attrEnvironment),
	)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//   COMPLIANCE METRIC HELPERS
// ────────────────────────────────────────────────────────────────────────────────
//

// ObserveComplianceLookupLatency records latency for a compliance lookup backend.
func ObserveComplianceLookupLatency(ctx context.Context, backend string, start time.Time) {
	if complianceLookupLatency == nil {
		return
	}

	d := time.Since(start)

	complianceLookupLatency.Record(
		ctx,
		float64(d.Milliseconds()),
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("backend", backend),
		),
	)
}

// IncrementComplianceRuleHit increments rule hit counter.
func IncrementComplianceRuleHit(ctx context.Context, ruleID, ruleType string, flag bool, sourceSystem string) {
	if complianceRuleHits == nil {
		return
	}

	complianceRuleHits.Add(
		ctx,
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("rule_id", ruleID),
			attribute.String("rule_type", ruleType),
			attribute.Bool("flag", flag),
			attribute.String("source_system", sourceSystem),
		),
	)
}

// IncrementComplianceRuleMiss increments rule miss counter.
func IncrementComplianceRuleMiss(ctx context.Context, entityType, sourceSystem string) {
	if complianceRuleMisses == nil {
		return
	}

	complianceRuleMisses.Add(
		ctx,
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("entity_type", entityType),
			attribute.String("source_system", sourceSystem),
		),
	)
}

// IncrementComplianceFallback increments fallback counter.
func IncrementComplianceFallback(ctx context.Context, entityType, sourceSystem string) {
	if complianceFallbacks == nil {
		return
	}

	complianceFallbacks.Add(
		ctx,
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("entity_type", entityType),
			attribute.String("source_system", sourceSystem),
		),
	)
}

// IncrementComplianceError increments compliance error counter.
func IncrementComplianceError(ctx context.Context, errorType string) {
	if complianceErrors == nil {
		return
	}

	complianceErrors.Add(
		ctx,
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("error_type", errorType),
		),
	)
}

// IncrementComplianceBackendFailure increments backend failure counter.
func IncrementComplianceBackendFailure(ctx context.Context, backend, reason string) {
	if complianceBackendFailures == nil {
		return
	}

	complianceBackendFailures.Add(
		ctx,
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("backend", backend),
			attribute.String("reason", reason),
		),
	)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//   RESILIENCE METRIC HELPERS
// ────────────────────────────────────────────────────────────────────────────────
//

// RecordRetryAttempt increments retry attempts counter.
func RecordRetryAttempt(ctx context.Context, stage, reason string) {
	if retryAttempts == nil {
		return
	}

	retryAttempts.Add(
		ctx,
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("stage", stage),
			attribute.String("reason", reason),
		),
	)
}

// RecordTimeout increments timeout counter.
func RecordTimeout(ctx context.Context, stage string) {
	if timeoutTotal == nil {
		return
	}

	timeoutTotal.Add(
		ctx,
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("stage", stage),
		),
	)
}

// RecordCircuitBreakerOpen increments circuit breaker open counter.
func RecordCircuitBreakerOpen(ctx context.Context, stage string) {
	if circuitBreakerOpen == nil {
		return
	}

	circuitBreakerOpen.Add(
		ctx,
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("stage", stage),
		),
	)
}

// RecordCircuitBreakerHalfOpen increments circuit breaker half-open counter.
func RecordCircuitBreakerHalfOpen(ctx context.Context, stage string) {
	if circuitBreakerHalfOpen == nil {
		return
	}

	circuitBreakerHalfOpen.Add(
		ctx,
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("stage", stage),
		),
	)
}

// RecordCircuitBreakerClosed increments circuit breaker closed counter.
func RecordCircuitBreakerClosed(ctx context.Context, stage string) {
	if circuitBreakerClosed == nil {
		return
	}

	circuitBreakerClosed.Add(
		ctx,
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("stage", stage),
		),
	)
}

// RecordBackpressureRejection increments backpressure rejection counter.
func RecordBackpressureRejection(ctx context.Context, stage string) {
	if backpressureRejections == nil {
		return
	}

	backpressureRejections.Add(
		ctx,
		1,
		metric.WithAttributes(
			attrService,
			attrEnvironment,
			attribute.String("stage", stage),
		),
	)
}
