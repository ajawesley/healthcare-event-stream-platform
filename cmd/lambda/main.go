package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/ajawes/hesp/internal/observability"
)

func main() {
	// -----------------------------------------
	// Initialize Observability for Lambda
	// -----------------------------------------
	observability.NewLogger("hesp-lambda", "hesp-lambda")
	observability.InitMetrics("hesp-lambda", "dev")
	observability.InitTracing("hesp-lambda", "v1.0.0", "dev")

	// Wrap handler with tracing + structured logging
	lambda.Start(observability.LambdaHandler(handler))
}

func handler(ctx context.Context, s3Event events.S3Event) (string, error) {
	log := observability.WithTrace(ctx)

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		observability.Error(ctx, "aws_config_failed", err, "aws_config_error", "failed to load AWS config")
		return "", fmt.Errorf("aws config: %w", err)
	}

	glueClient := glue.NewFromConfig(awsCfg)

	// Extract S3 event details
	record := s3Event.Records[0]
	bucket := record.S3.Bucket.Name
	key := record.S3.Object.Key

	inputPath := fmt.Sprintf("s3://%s/%s", bucket, key)

	log.Info("lambda_s3_event_received",
		zap.String("bucket", bucket),
		zap.String("key", key),
		zap.String("input_path", inputPath),
	)

	// Glue job name from env var (best practice)
	jobName := os.Getenv("GLUE_JOB_NAME")
	if jobName == "" {
		jobName = "hesp-dev-job"
	}

	// Output + error paths from env
	outputPath := os.Getenv("OUTPUT_BASE_PATH")
	if outputPath == "" {
		return "", fmt.Errorf("missing OUTPUT_BASE_PATH env var")
	}

	errorPath := os.Getenv("ERROR_PATH")
	if errorPath == "" {
		return "", fmt.Errorf("missing ERROR_PATH env var")
	}

	// -----------------------------------------
	// Start OTEL span for Glue job invocation
	// -----------------------------------------
	tracer := otel.Tracer("hesp-lambda")
	ctx, span := tracer.Start(ctx, "start_glue_job")
	defer span.End()

	sc := span.SpanContext()
	traceID := sc.TraceID().String()
	spanID := sc.SpanID().String()

	// -----------------------------------------
	// Start Glue job with trace propagation
	// -----------------------------------------
	_, err = glueClient.StartJobRun(ctx, &glue.StartJobRunInput{
		JobName: aws.String(jobName),
		Arguments: map[string]string{
			"--JOB_NAME":         jobName,
			"--input_path":       inputPath,
			"--output_base_path": outputPath,
			"--error_path":       errorPath,
			"--trace_id":         traceID,
			"--span_id":          spanID,
		},
	})
	if err != nil {
		observability.Error(ctx, "glue_start_failed", err, "glue_error", "failed to start Glue job",
			zap.String("job_name", jobName),
			zap.String("input_path", inputPath),
		)
		return "", fmt.Errorf("start glue: %w", err)
	}

	log.Info("glue_job_started",
		zap.String("job_name", jobName),
		zap.String("input_path", inputPath),
		zap.String("trace_id", traceID),
		zap.String("span_id", spanID),
	)

	return "ok", nil
}
