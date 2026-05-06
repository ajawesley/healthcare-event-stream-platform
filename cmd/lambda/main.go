package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/glue"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, s3Event events.S3Event) error {
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("aws config: %w", err)
	}

	glueClient := glue.NewFromConfig(awsCfg)

	// Extract S3 event details
	record := s3Event.Records[0]
	bucket := record.S3.Bucket.Name
	key := record.S3.Object.Key

	inputPath := fmt.Sprintf("s3://%s/%s", bucket, key)
	log.Printf("Lambda triggered for %s", inputPath)

	// Glue job name from env var (best practice)
	jobName := os.Getenv("GLUE_JOB_NAME")
	if jobName == "" {
		jobName = "hesp-dev-job"
	}

	// Start Glue job with the EXACT file path
	_, err = glueClient.StartJobRun(ctx, &glue.StartJobRunInput{
		JobName: aws.String(jobName),
		Arguments: map[string]string{
			"--input_path": inputPath,
		},
	})
	if err != nil {
		return fmt.Errorf("start glue: %w", err)
	}

	log.Printf("Glue job %s started for %s", jobName, key)
	return nil
}
