func (d *S3Dispatcher) Dispatch(event *models.CanonicalEvent, env api.Envelope, raw []byte) error {
	ctx := context.Background()
	start := time.Now()

	// Build S3 key
	fullKey := fmt.Sprintf(
		"%s/%s/%s/%s.json",
		d.prefix,
		env.SourceSystem,
		env.EventType,
		event.EventID,
	)

	// Build the full S3 object payload
	obj := map[string]any{
		"envelope":        env,
		"canonical_event": event,
		"raw":             string(raw),
		"dispatched_at":   time.Now().UTC().Format(time.RFC3339),
	}

	// Marshal the full object
	payload, err := json.Marshal(obj)
	if err != nil {
		d.logger.Error("json_marshal_failed",
			slog.String("bucket", d.bucket),
			slog.String("key", fullKey),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("marshal s3 object: %w", err)
	}

	// Compute MD5 over EXACT payload bytes
	sum := md5.Sum(payload)
	md5b64 := base64.StdEncoding.EncodeToString(sum[:])

	d.logger.Info("s3_dispatch_start",
		slog.String("bucket", d.bucket),
		slog.String("key", fullKey),
		slog.String("event_id", event.EventID),
		slog.String("event_type", env.EventType),
		slog.String("source_system", env.SourceSystem),
		slog.String("kms_key_arn", d.kmsKeyARN),
		slog.String("md5", md5b64),
		slog.Int("payload_bytes", len(payload)),
	)

	// Write to S3 with SSE-KMS
	_, err = d.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               aws.String(d.bucket),
		Key:                  aws.String(fullKey),
		Body:                 bytes.NewReader(payload),
		ContentMD5:           aws.String(md5b64),
		ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		SSEKMSKeyId:          aws.String(d.kmsKeyARN),
		ContentType:          aws.String("application/json"),
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
