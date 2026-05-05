package dispatcher

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"time"

	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

type S3Dispatcher struct {
	client    *s3.Client
	bucket    string
	prefix    string
	kmsKeyARN string
	logger    *slog.Logger
}

func NewS3Dispatcher(client *s3.Client, bucket, prefix, kmsKeyARN string, logger *slog.Logger) *S3Dispatcher {
	return &S3Dispatcher{
		client:    client,
		bucket:    bucket,
		prefix:    prefix,
		kmsKeyARN: kmsKeyARN,
		logger:    logger,
	}
}

func (d *S3Dispatcher) Dispatch(event *models.CanonicalEvent, env api.Envelope, raw []byte) error {
	ctx := context.Background()
	start := time.Now()

	// Build S3 key using event metadata
	fullKey := fmt.Sprintf(
		"%s/%s/%s/%s.json",
		d.prefix,
		env.SourceSystem,
		env.EventType,
		event.EventID,
	)

	// Compute MD5 checksum
	sum := md5.Sum(raw)
	md5b64 := base64.StdEncoding.EncodeToString(sum[:])

	d.logger.Info("s3_dispatch_start",
		slog.String("bucket", d.bucket),
		slog.String("key", fullKey),
		slog.String("event_id", event.EventID),
		slog.String("event_type", env.EventType),
		slog.String("source_system", env.SourceSystem),
		slog.String("kms_key_arn", d.kmsKeyARN),
		slog.String("md5", md5b64),
		slog.Int("raw_bytes", len(raw)),
	)

	// Write to S3 with SSE-KMS
	_, err := d.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               aws.String(d.bucket),
		Key:                  aws.String(fullKey),
		Body:                 bytes.NewReader(raw),
		ContentMD5:           aws.String(md5b64),
		ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		SSEKMSKeyId:          aws.String(d.kmsKeyARN),
	})
	if err != nil {
		d.logger.Error("s3_dispatch_failed",
			slog.String("bucket", d.bucket),
			slog.String("key", fullKey),
			slog.String("error", err.Error()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
		return fmt.Errorf("s3 dispatcher failed: %w", err)
	}

	d.logger.Info("s3_dispatch_success",
		slog.String("bucket", d.bucket),
		slog.String("key", fullKey),
		slog.String("event_id", event.EventID),
		slog.String("md5", md5b64),
		slog.Int64("duration_ms", time.Since(start).Milliseconds()),
	)

	return nil
}
