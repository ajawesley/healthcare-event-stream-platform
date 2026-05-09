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

	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
	"github.com/aws/aws-xray-sdk-go/xray"

	"go.uber.org/zap"

	"github.com/ajawes/hesp/internal/observability"
)

func init() {
	// Enable AWS X-Ray for AWS SDK calls
	xray.Configure(xray.Config{
		LogLevel: "info",
	})
}

func main() {
	// -----------------------------------------
	// Initialize logging + metrics
	// -----------------------------------------
	observability.NewLogger("hesp-lambda", "hesp-lambda")
	observability.InitMetrics("hesp-lambda", "dev")

	// -----------------------------------------
	// Start Lambda handler (no OTEL wrapper)
	// -----------------------------------------
	lambda.Start(handler)
}

func handler(ctx context.Context, s3Event events.S3Event) (string, error) {
	// -----------------------------------------
	// Start X-Ray segment
	// -----------------------------------------
	ctx, seg := xray.BeginSegment(ctx, "hesp-lambda-handler")
	defer seg.Close(nil)

	log := observability.WithTrace(ctx)

	// -----------------------------------------
	// Extract S3 event
	// -----------------------------------------
	if len(s3Event.Records) == 0 {
		return "", fmt.Errorf("no S3 records in event")
	}

	record := s3Event.Records[0]
	bucket := record.S3.Bucket.Name
	key := record.S3.Object.Key
	inputPath := fmt.Sprintf("s3://%s/%s", bucket, key)

	// -----------------------------------------
	// Extract trace ID from X-Ray context
	// -----------------------------------------
	traceID := xray.TraceID(ctx)

	// Generate a new span ID
	spanID := xray.NewSegmentID()

	log.Info("lambda_s3_event_received",
		zap.String("bucket", bucket),
		zap.String("key", key),
		zap.String("input_path", inputPath),
		zap.String("trace_id", traceID),
		zap.String("span_id", spanID),
	)

	// -----------------------------------------
	// Load AWS config with X-Ray instrumentation
	// -----------------------------------------
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		observability.Error(ctx, "aws_config_failed", err,
			"aws_config_error", "failed to load AWS config",
			zap.String("bucket", bucket),
			zap.String("key", key),
		)
		return "", fmt.Errorf("aws config: %w", err)
	}

	// Instrument AWS SDK v2 with X-Ray
	awsv2.AWSV2Instrumentor(&awsCfg.APIOptions)

	glueClient := glue.NewFromConfig(awsCfg)

	// -----------------------------------------
	// Resolve Glue job name + paths
	// -----------------------------------------
	jobName := os.Getenv("GLUE_JOB_NAME")
	if jobName == "" {
		jobName = "hesp-dev-job"
	}

	outputPath := os.Getenv("OUTPUT_BASE_PATH")
	if outputPath == "" {
		return "", fmt.Errorf("missing OUTPUT_BASE_PATH env var")
	}

	errorPath := os.Getenv("ERROR_PATH")
	if errorPath == "" {
		return "", fmt.Errorf("missing ERROR_PATH env var")
	}

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
		observability.Error(ctx, "glue_start_failed", err,
			"glue_error", "failed to start Glue job",
			zap.String("job_name", jobName),
			zap.String("input_path", inputPath),
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
		)
		return "", fmt.Errorf("start glue: %w", err)
	}

	log.Info("glue_job_started",
		zap.String("job_name", jobName),
		zap.String("input_path", inputPath),
		zap.String("output_base_path", outputPath),
		zap.String("error_path", errorPath),
		zap.String("trace_id", traceID),
		zap.String("span_id", spanID),
	)

	return "ok", nil
}
