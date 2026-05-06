package observability

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger       *zap.Logger
	serviceName  string
	sourceSystem string
)

// -----------------------------------------------------------------------------
// Initialization
// -----------------------------------------------------------------------------

// NewLogger initializes the global logger with service metadata.
func NewLogger(svc string, source string) {
	serviceName = svc
	sourceSystem = source

	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:    "message",
			LevelKey:      "level",
			TimeKey:       "produced_at",
			EncodeTime:    zapcore.ISO8601TimeEncoder,
			EncodeLevel:   zapcore.LowercaseLevelEncoder,
			CallerKey:     "caller",
			EncodeCaller:  zapcore.ShortCallerEncoder,
			StacktraceKey: "stacktrace",
		},
	}

	l, err := cfg.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	logger = l
}

// -----------------------------------------------------------------------------
// Context‑Aware Logger Enrichment
// -----------------------------------------------------------------------------

// WithTrace enriches logs with trace_id, span_id, event_id, produced_at, and metadata.
func WithTrace(ctx context.Context) *zap.Logger {
	if logger == nil {
		panic("logger not initialized — call NewLogger() in main()")
	}

	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()

	return logger.With(
		zap.String("trace_id", sc.TraceID().String()),
		zap.String("span_id", sc.SpanID().String()),
		zap.String("event_id", uuid.New().String()),
		zap.String("source_system", sourceSystem),
		zap.String("service", serviceName),
		zap.String("format", "json"),
		zap.String("produced_at", time.Now().UTC().Format(time.RFC3339Nano)),
	)
}

// -----------------------------------------------------------------------------
// Logging Helpers
// -----------------------------------------------------------------------------

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	WithTrace(ctx).Info(msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	WithTrace(ctx).Warn(msg, fields...)
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	WithTrace(ctx).Debug(msg, fields...)
}

func Error(ctx context.Context, msg string, err error, code string, reason string, fields ...zap.Field) {
	all := append(fields,
		zap.String("error_code", code),
		zap.String("error_reason", reason),
		zap.String("error_message", err.Error()),
	)
	WithTrace(ctx).Error(msg, all...)
}

// -----------------------------------------------------------------------------
// Flush on shutdown
// -----------------------------------------------------------------------------

func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}
