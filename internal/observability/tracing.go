package observability

import (
    "context"
    "fmt"
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
        endpoint = "adot:4317"
    }
    fmt.Printf("[OTEL] Using OTLP endpoint: %s\n", endpoint)

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
    // gRPC connection to OTEL collector
    // -----------------------------------------
    fmt.Printf("[OTEL] Attempting gRPC dial to collector...\n")

    conn, err := grpc.DialContext(
        ctx,
        endpoint,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
        grpc.WithTimeout(5*time.Second),
    )
    if err != nil {
        fmt.Printf("[OTEL ERROR] grpc.DialContext failed: %v\n", err)
        return func(context.Context) error { return nil }
    }

    fmt.Printf("[OTEL] gRPC dial successful.\n")

    exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
    if err != nil {
        fmt.Printf("[OTEL ERROR] otlptracegrpc.New failed: %v\n", err)
        return func(context.Context) error { return nil }
    }

    fmt.Printf("[OTEL] OTLP gRPC exporter initialized.\n")

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
        if err := tp.Shutdown(ctx); err != nil {
            fmt.Printf("[OTEL ERROR] Shutdown failed: %v\n", err)
            return err
        }
        fmt.Printf("[OTEL] Shutdown complete.\n")
        return nil
    }
}
