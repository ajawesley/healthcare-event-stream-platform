package observability

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InitTracing initializes OpenTelemetry tracing and returns a shutdown function.
// This works for ECS, Lambda, and local dev.
func InitTracing(serviceName, version, environment string) func(ctx context.Context) error {
	ctx := context.Background()

	// -----------------------------------------
	// OTLP endpoint (env‑driven)
	// -----------------------------------------
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		// Local default — ECS/Lambda MUST override this
		endpoint = "adot:4317"
	}

	// -----------------------------------------
	// Cloud region (env‑driven)
	// -----------------------------------------
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if region == "" {
		region = "us-east-1"
	}

	// -----------------------------------------
	// gRPC connection to OTEL collector
	// -----------------------------------------
	conn, err := grpc.DialContext(
		ctx,
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		// Fail open — tracing disabled but service still runs
		return func(context.Context) error { return nil }
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return func(context.Context) error { return nil }
	}

	// -----------------------------------------
	// Resource attributes (critical for Datadog/Honeycomb/X-Ray)
	// -----------------------------------------
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(version),
			semconv.DeploymentEnvironmentKey.String(environment),
			semconv.CloudRegionKey.String(region),
		),
	)
	if err != nil {
		return func(context.Context) error { return nil }
	}

	// -----------------------------------------
	// Tracer provider with batch processor
	// -----------------------------------------
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	// -----------------------------------------
	// Return shutdown function
	// -----------------------------------------
	return func(ctx context.Context) error {
		// Flush spans
		if err := tp.Shutdown(ctx); err != nil {
			return err
		}
		return nil
	}
}
