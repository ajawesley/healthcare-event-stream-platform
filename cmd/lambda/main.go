package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/aws/aws-sdk-go-v2/service/glue/types"
)

type ScheduledEvent struct {
	Source string `json:"source"`
}

var glueClient *glue.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}
	glueClient = glue.NewFromConfig(cfg)
}

func handler(ctx context.Context, raw json.RawMessage) error {
	// Try S3 event first
	var s3Event events.S3Event
	if err := json.Unmarshal(raw, &s3Event); err == nil && len(s3Event.Records) > 0 {
		return handleS3Event(ctx, s3Event)
	}

	// Try scheduled EventBridge event
	var sched ScheduledEvent
	if err := json.Unmarshal(raw, &sched); err == nil && sched.Source == "aws.events" {
		return handleScheduledEvent(ctx)
	}

	// Unknown event type: log and return nil to avoid noisy retries
	log.Printf("Received unsupported event type payload: %s", string(raw))
	return nil
}

func handleS3Event(ctx context.Context, event events.S3Event) error {
	log.Printf("Received S3 event with %d record(s)", len(event.Records))

	for _, record := range event.Records {
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key

		log.Printf("Processing S3 record: bucket=%s key=%s", bucket, key)

		dir := path.Dir(key)
		var keyPrefix string
		if dir == "." {
			keyPrefix = ""
		} else {
			// Ensure trailing slash
			if strings.HasSuffix(dir, "/") {
				keyPrefix = dir
			} else {
				keyPrefix = dir + "/"
			}
		}

		inputPath := fmt.Sprintf("s3://%s/%s", bucket, keyPrefix)
		outputPath := fmt.Sprintf("s3://%s/output/", bucket)
		errorPath := fmt.Sprintf("s3://%s/errors/", bucket)

		if err := startGlueJob(ctx, inputPath, outputPath, errorPath); err != nil {
			// If Glue is already running, log and continue instead of failing the whole batch
			if isConcurrentRunsExceeded(err) {
				log.Printf("Glue job already running for inputPath=%s; skipping: %v", inputPath, err)
				continue
			}
			// For other errors, fail so S3 can retry this batch
			return fmt.Errorf("failed to start Glue job for key=%s: %w", key, err)
		}
	}

	return nil
}

func handleScheduledEvent(ctx context.Context) error {
	log.Printf("Received scheduled EventBridge trigger")

	postingDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")

	rawBucket := os.Getenv("RAW_BUCKET")
	if rawBucket == "" {
		return fmt.Errorf("RAW_BUCKET environment variable not set")
	}

	inputPath := fmt.Sprintf("s3://%s/postingDate=%s/", rawBucket, postingDate)
	outputPath := fmt.Sprintf("s3://%s/output/", rawBucket)
	errorPath := fmt.Sprintf("s3://%s/errors/", rawBucket)

	if err := startGlueJob(ctx, inputPath, outputPath, errorPath); err != nil {
		if isConcurrentRunsExceeded(err) {
			log.Printf("Glue job already running for scheduled postingDate=%s; skipping: %v", postingDate, err)
			return nil
		}
		return fmt.Errorf("failed to start Glue job for scheduled run: %w", err)
	}

	return nil
}

func startGlueJob(ctx context.Context, inputPath, outputPath, errorPath string) error {
	jobName := os.Getenv("GLUE_JOB_NAME")
	if jobName == "" {
		return fmt.Errorf("GLUE_JOB_NAME environment variable not set")
	}

	log.Printf("Starting Glue job: %s", jobName)
	log.Printf("input_path=%s", inputPath)
	log.Printf("output_base_path=%s", outputPath)
	log.Printf("error_path=%s", errorPath)

	input := &glue.StartJobRunInput{
		JobName: aws.String(jobName),
		Arguments: map[string]string{
			"--input_path":       inputPath,
			"--output_base_path": outputPath,
			"--error_path":       errorPath,
		},
	}

	out, err := glueClient.StartJobRun(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to start Glue job: %w", err)
	}

	log.Printf("Glue job started successfully: jobRunId=%s", aws.ToString(out.JobRunId))
	return nil
}

func isConcurrentRunsExceeded(err error) bool {
	var concErr *types.ConcurrentRunsExceededException
	if errors.As(err, &concErr) {
		return true
	}
	// Fallback: some SDK/regions may surface this as a generic error with a message
	if strings.Contains(err.Error(), "ConcurrentRunsExceededException") {
		return true
	}
	return false
}

func main() {
	lambda.Start(handler)
}
