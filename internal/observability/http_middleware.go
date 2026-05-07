package observability

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// -----------------------------------------------------------------------------
// ECS Observability Middleware
// -----------------------------------------------------------------------------

func ObservabilityMiddleware(next http.Handler) http.Handler {
	tracer := otel.Tracer("hesp-ecs")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Normalize route (prevents cardinality explosion)
		route := NormalizeRoute(r.URL.Path)

		// Create a new span for the request
		ctx, span := tracer.Start(
			r.Context(),
			"http.request",
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", route),
				attribute.String("http.target", r.URL.String()),
				attribute.String("http.user_agent", r.UserAgent()),
			),
		)
		defer span.End()

		// Wrap ResponseWriter to capture status code
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

		// Log request started
		Info(ctx, "request_started",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("route", route),
			zap.String("remote_addr", r.RemoteAddr),
		)

		// Panic recovery
		defer func() {
			if rec := recover(); rec != nil {
				Error(ctx, "panic_recovered",
					nil,
					"panic",
					"handler panic",
					zap.Any("panic_value", rec),
				)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()

		// Call next handler
		next.ServeHTTP(rec, r.WithContext(ctx))

		// Duration
		latency := time.Since(start)

		// Emit metrics
		RecordRequestCount(r.Method, rec.status)
		RecordRequestLatency(r.Method, rec.status, latency)

		if rec.status >= 500 {
			RecordHttpError(r.Method, rec.status)
		}

		// Enrich span
		span.SetAttributes(
			attribute.Int("http.status_code", rec.status),
			attribute.Float64("http.latency_ms", float64(latency.Milliseconds())),
		)

		// Log request completed
		Info(ctx, "request_completed",
			zap.Int("status", rec.status),
			zap.Duration("latency_ms", latency),
		)
	})
}
