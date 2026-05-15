package observability

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc"
)

// InitTracing initializes OpenTelemetry tracing and returns a shutdown function.
func InitTracing(serviceName, version, environment string) func(ctx context.Context) error {
	ctx := context.Background()

	// -----------------------------------------
	// OTLP endpoint (env‑driven, gRPC)
	// -----------------------------------------
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		// Default to ADOT sidecar OTLP gRPC
		endpoint = "127.0.0.1:4317"
	}
	fmt.Printf("[OTEL] Using OTLP gRPC endpoint: %s\n", endpoint)

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
	fmt.Printf("[OTEL] Using cloud region: %s\n", region)

	// -----------------------------------------
	// OTLP gRPC exporter client
	// -----------------------------------------
	fmt.Printf("[OTEL] Initializing OTLP gRPC exporter...\n")

	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		fmt.Printf("[OTEL ERROR] otlptrace.New failed: %v\n", err)
		return func(context.Context) error { return nil }
	}

	fmt.Printf("[OTEL] OTLP gRPC exporter initialized.\n")

	// -----------------------------------------
	// Resource attributes
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
		fmt.Printf("[OTEL ERROR] resource.New failed: %v\n", err)
		return func(context.Context) error { return nil }
	}

	fmt.Printf("[OTEL] Resource attributes initialized.\n")

	// -----------------------------------------
	// Tracer provider with batch processor
	// -----------------------------------------
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	fmt.Printf("[OTEL] Tracer provider successfully installed.\n")

	// -----------------------------------------
	// Return shutdown function
	// -----------------------------------------
	return func(ctx context.Context) error {
		fmt.Printf("[OTEL] Shutting down tracer provider...\n")

		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := tp.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("[OTEL ERROR] Shutdown failed: %v\n", err)
			return err
		}

		fmt.Printf("[OTEL] Shutdown complete.\n")
		return nil
	}
}
