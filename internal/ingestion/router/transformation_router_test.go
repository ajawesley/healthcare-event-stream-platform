package router

import (
	"context"
	"testing"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/observability"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// ------------------------------------------------------------
// Observability initialization for tests
// ------------------------------------------------------------
func init() {
	observability.NewLogger("hesp-ecs", "test")
	observability.InitMetrics("hesp-ecs", "test")
	otel.SetTracerProvider(trace.NewNoopTracerProvider())
}

func TestTransformationRouter(t *testing.T) {
	r := NewTransformationRouter()

	tests := []struct {
		name      string
		format    config.Format
		expectErr bool
	}{
		{"hl7", config.FormatHL7, false},
		{"x12", config.FormatX12, false},
		{"fhir", config.FormatFHIR, false},
		{"generic", config.FormatGeneric, false},
		{"unknown", config.Format("weird"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := r.TransformerFor(context.Background(), tt.format)

			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
