package dispatcher

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/observability"
	"github.com/ajawes/hesp/internal/resilience"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type S3Dispatcher struct {
	client    *s3.Client
	bucket    string
	prefix    string
	kmsKeyARN string
}

func NewS3Dispatcher(client *s3.Client, bucket, prefix, kmsKeyARN string) *S3Dispatcher {
	return &S3Dispatcher{
		client:    client,
		bucket:    bucket,
		prefix:    prefix,
		kmsKeyARN: kmsKeyARN,
	}
}

// -----------------------------------------------------------------------------
// Dispatch writes the canonical event + envelope + raw payload to S3.
// Includes:
//   - trace_id / span_id
//   - lineage stages
//   - compliance metadata (inside canonical_event)
//   - transmission_timestamp
//   - dispatched_at
//
// -----------------------------------------------------------------------------
func (d *S3Dispatcher) Dispatch(ctx context.Context, event *models.CanonicalEvent, env api.Envelope, raw []byte) error {
	start := time.Now()

	// Use existing span if present
	tr := trace.SpanFromContext(ctx).TracerProvider().Tracer("hesp-ecs")
	ctx, span := tr.Start(ctx, "s3.dispatch")
	defer span.End()

	log := observability.WithTrace(ctx)

	// Extract trace + span IDs for payload
	sc := span.SpanContext()
	traceID := sc.TraceID().String()
	spanID := sc.SpanID().String()

	fullKey := fmt.Sprintf(
		"%s/%s/%s/%s.json",
		d.prefix,
		env.SourceSystem,
		env.EventType,
		event.EventID,
	)

	log.Info("s3_dispatch_initiated",
		zap.String("bucket", d.bucket),
		zap.String("key", fullKey),
		zap.String("event_id", event.EventID),
		zap.String("event_type", env.EventType),
		zap.String("source_system", env.SourceSystem),
		zap.String("trace_id", traceID),
		zap.String("span_id", spanID),
		zap.String("raw_bytes", strconv.Quote(string(raw))),
	)

	// Transmission timestamp BEFORE payload build
	transmissionTime := time.Now().UTC()

	// -------------------------------------------------------------------------
	// Lineage stages
	// -------------------------------------------------------------------------
	var lineageStages []observability.LineageStage
	if lineage := observability.GetLineage(ctx); lineage != nil {
		lineageStages = lineage.Stages()
	}

	// -------------------------------------------------------------------------
	// Build S3 object payload
	// -------------------------------------------------------------------------
	obj := map[string]any{
		"envelope":               env,
		"canonical_event":        event,
		"raw":                    string(raw),
		"lineage":                lineageStages,
		"trace_id":               traceID,
		"span_id":                spanID,
		"transmission_timestamp": transmissionTime.Format(time.RFC3339Nano),
		"dispatched_at":          time.Now().UTC().Format(time.RFC3339Nano),
	}

	payload, err := json.Marshal(obj)
	if err != nil {
		observability.Error(ctx, "json_marshal_failed", err, "marshal_error", "failed to marshal S3 payload",
			zap.String("bucket", d.bucket),
			zap.String("key", fullKey),
		)
		span.RecordError(err)
		return fmt.Errorf("marshal s3 object: %w", err)
	}

	// Compute MD5 checksum
	sum := md5.Sum(payload)
	md5b64 := base64.StdEncoding.EncodeToString(sum[:])

	log.Info("s3_dispatch_start",
		zap.String("bucket", d.bucket),
		zap.String("key", fullKey),
		zap.String("event_id", event.EventID),
		zap.String("trace_id", traceID),
		zap.String("span_id", spanID),
		zap.String("md5", md5b64),
		zap.Int("payload_bytes", len(payload)),
	)

	// -------------------------------------------------------------------------
	// Write to S3 with resiliency
	// -------------------------------------------------------------------------
	err = resilience.DoWithFallback(ctx, resilience.Dependency("s3"), func(ctx context.Context) error {
		s3Start := time.Now()

		_, putErr := d.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:               aws.String(d.bucket),
			Key:                  aws.String(fullKey),
			Body:                 bytes.NewReader(payload),
			ContentMD5:           aws.String(md5b64),
			ServerSideEncryption: types.ServerSideEncryptionAwsKms,
			SSEKMSKeyId:          aws.String(d.kmsKeyARN),
			ContentType:          aws.String("application/json"),
		})

		s3Duration := time.Since(s3Start)
		observability.ObserveS3PutLatency(ctx, d.bucket, fullKey, s3Duration, putErr == nil)

		if putErr != nil {
			observability.Error(ctx, "s3_dispatch_failed", putErr, "s3_error", "failed to write object to S3",
				zap.String("bucket", d.bucket),
				zap.String("key", fullKey),
				zap.String("trace_id", traceID),
				zap.String("span_id", spanID),
				zap.Int64("duration_ms", time.Since(start).Milliseconds()),
			)
			span.RecordError(putErr)
			return putErr
		}

		return nil
	}, func(ctx context.Context, fbErr error) error {
		observability.Error(ctx, "s3_dispatch_resiliency_fallback", fbErr, "s3_fallback", "s3 dispatch failed after resiliency",
			zap.String("bucket", d.bucket),
			zap.String("key", fullKey),
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
		)
		span.RecordError(fbErr)
		return fbErr
	})

	if err != nil {
		return fmt.Errorf("s3 dispatcher failed: %w", err)
	}

	// -------------------------------------------------------------------------
	// Mark lineage stage: written
	// -------------------------------------------------------------------------
	if lineage := observability.GetLineage(ctx); lineage != nil {
		stageStart := time.Now()
		lineage.MarkStage("written")

		observability.Info(ctx, "lineage_stage_written",
			zap.String("event_id", lineage.EventID),
			zap.String("trace_id", lineage.TraceID),
			zap.Any("stages", lineage.Stages()),
		)

		observability.ObserveLineageLatency(ctx, "written", stageStart)
	}

	log.Info("s3_dispatch_success",
		zap.String("bucket", d.bucket),
		zap.String("key", fullKey),
		zap.String("event_id", event.EventID),
		zap.String("trace_id", traceID),
		zap.String("span_id", spanID),
		zap.String("md5", md5b64),
		zap.Int64("duration_ms", time.Since(start).Milliseconds()),
	)

	span.SetAttributes(
		attribute.String("s3.bucket", d.bucket),
		attribute.String("s3.key", fullKey),
		attribute.String("event.id", event.EventID),
		attribute.String("event.type", env.EventType),
		attribute.String("source_system", env.SourceSystem),
		attribute.String("trace_id", traceID),
		attribute.String("span_id", spanID),
		attribute.Int("payload_bytes", len(payload)),
	)

	return nil
}
