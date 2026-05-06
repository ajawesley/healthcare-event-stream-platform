package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// LambdaHandler wraps AWS Lambda handlers with full observability.
// Generic version: supports strongly typed event + response.
func LambdaHandler[TIn any, TOut any](
	handler func(ctx context.Context, event TIn) (TOut, error),
) func(ctx context.Context, event TIn) (TOut, error) {

	tracer := otel.Tracer("hesp-lambda")

	return func(ctx context.Context, event TIn) (TOut, error) {
		start := time.Now()

		// Extract AWS Lambda context
		lc, _ := lambdacontext.FromContext(ctx)
		awsReqID := ""
		if lc != nil {
			awsReqID = lc.AwsRequestID
		}

		// Start span
		ctx, span := tracer.Start(
			ctx,
			"lambda.invocation",
			trace.WithAttributes(
				attribute.String("aws.request_id", awsReqID),
				attribute.String("lambda.event_type", fmt.Sprintf("%T", event)),
			),
		)
		defer span.End()

		// Log invocation started
		Info(ctx, "lambda_invocation_started",
			zap.String("aws_request_id", awsReqID),
			zap.String("event_type", fmt.Sprintf("%T", event)),
		)

		// Panic recovery
		defer func() {
			if rec := recover(); rec != nil {
				Error(ctx, "panic_recovered",
					nil,
					"panic",
					"lambda handler panic",
					zap.Any("panic_value", rec),
				)
			}
		}()

		// Execute handler
		resp, err := handler(ctx, event)

		latency := time.Since(start)

		// Emit metrics
		RecordLambdaInvocation()
		RecordLambdaLatency(latency)
		if err != nil {
			RecordLambdaError()
		}

		// Log invocation completed
		if err != nil {
			Error(ctx, "lambda_invocation_failed",
				err,
				"lambda_error",
				"handler returned error",
				zap.Duration("latency_ms", latency),
				zap.String("aws_request_id", awsReqID),
			)
		} else {
			Info(ctx, "lambda_invocation_completed",
				zap.Duration("latency_ms", latency),
				zap.String("aws_request_id", awsReqID),
			)
		}

		return resp, err
	}
}
