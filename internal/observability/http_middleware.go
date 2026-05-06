package observability

import (
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// ObservabilityMiddleware wraps all HTTP requests with tracing + structured logging.
func ObservabilityMiddleware(next http.Handler) http.Handler {
	tracer := otel.Tracer("hesp-ecs-http")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ctx, span := tracer.Start(r.Context(), "http.request")
		defer span.End()

		log := WithTrace(ctx)

		rr := newResponseRecorder(w)

		// request_started
		log.Info("request_started",
			zap.String("http_method", r.Method),
			zap.String("http_path", r.URL.Path),
			zap.String("http_host", r.Host),
			zap.String("http_user_agent", r.UserAgent()),
		)

		defer func() {
			if rec := recover(); rec != nil {
				err := fmt.Errorf("panic: %v", rec)
				span.RecordError(err)

				http.Error(rr, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				Error(ctx, "request_panic", err, "panic", "handler panicked",
					zap.String("http_method", r.Method),
					zap.String("http_path", r.URL.Path),
				)
			}
		}()

		next.ServeHTTP(rr, r.WithContext(ctx))

		duration := time.Since(start)

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.target", r.URL.Path),
			attribute.Int("http.status_code", rr.statusCode),
			attribute.Int("http.response_size", rr.bytes),
			attribute.Float64("http.server.duration_ms", float64(duration.Milliseconds())),
		)

		// request_completed
		log.Info("request_completed",
			zap.String("http_method", r.Method),
			zap.String("http_path", r.URL.Path),
			zap.Int("http_status_code", rr.statusCode),
			zap.Int("response_bytes", rr.bytes),
			zap.Int64("duration_ms", duration.Milliseconds()),
		)
	})
}
