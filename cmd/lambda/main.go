package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, s3Event events.S3Event) error {
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("aws config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg)
	glueClient := glue.NewFromConfig(awsCfg)

	record := s3Event.Records[0]
	bucket := record.S3.Bucket.Name
	key := record.S3.Object.Key

	log.Printf("Lambda triggered for s3://%s/%s", bucket, key)

	obj, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("get object: %w", err)
	}
	defer obj.Body.Close()

	var event map[string]any
	if err := json.NewDecoder(obj.Body).Decode(&event); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}

	log.Printf("Event read: %v", event)

	_, err = glueClient.StartJobRun(ctx, &glue.StartJobRunInput{
		JobName: aws.String("hesp-demo-job"),
		Arguments: map[string]string{
			"--input_s3_key": fmt.Sprintf("s3://%s/%s", bucket, key),
		},
	})
	if err != nil {
		return fmt.Errorf("start glue: %w", err)
	}

	log.Printf("Glue job started for %s", key)
	return nil
}
