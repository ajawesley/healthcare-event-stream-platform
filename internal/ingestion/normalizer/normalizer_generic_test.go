package normalizer

import (
	"context"
	"testing"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/observability"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace/noop"
)

// ------------------------------------------------------------
// Observability initialization for tests
// ------------------------------------------------------------
func init() {
	observability.NewLogger("hesp-ecs", "test")
	observability.InitMetrics("hesp-ecs", "test")
	otel.SetTracerProvider(noop.NewTracerProvider())
}

func TestGenericNormalizer(t *testing.T) {
	n := NewGenericNormalizer()

	tests := []struct {
		name      string
		raw       []byte
		env       api.Envelope
		expectErr bool
		verify    func(t *testing.T, format config.Format, fields map[string]any)
	}{
		{
			name: "Wraps raw payload into fields.raw",
			raw:  []byte(`{"a":1}`),
			env:  api.Envelope{EventID: "evt1", SourceSystem: "sys"},
			verify: func(t *testing.T, format config.Format, fields map[string]any) {
				if format != config.FormatGeneric {
					t.Fatalf("expected format=generic, got %s", format)
				}

				rawVal, ok := fields["raw"].(string)
				if !ok {
					t.Fatalf("expected raw field to be string, got %T", fields["raw"])
				}

				if rawVal != `{"a":1}` {
					t.Fatalf("expected raw payload preserved, got %s", rawVal)
				}
			},
		},
		{
			name:      "Nil raw returns error",
			raw:       nil,
			env:       api.Envelope{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// ⭐ UPDATED: Normalize now requires ctx
			ne, err := n.Normalize(context.Background(), tt.raw, tt.env)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.verify != nil {
				tt.verify(t, config.Format(ne.Format), ne.Fields)
			}
		})
	}
}
