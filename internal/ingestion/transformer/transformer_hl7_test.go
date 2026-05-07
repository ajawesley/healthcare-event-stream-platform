package transformer

import (
	"context"
	"testing"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
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

func TestHL7Transformer(t *testing.T) {
	tests := []struct {
		name      string
		ne        *models.NormalizedEvent
		env       api.Envelope
		expectErr bool
		verify    func(t *testing.T, ce *models.CanonicalEvent)
	}{
		{
			name: "Maps patient, encounter, observation",
			ne: &models.NormalizedEvent{
				Format: config.FormatHL7,
				Raw:    "raw",
				Fields: map[string]any{
					"pid.id":             "PAT123",
					"pid.first_name":     "John",
					"pid.last_name":      "Doe",
					"pv1.encounter_type": "I",
					"msh.message_type":   "ORU^R01",
				},
			},
			env: api.Envelope{EventID: "evt1", SourceSystem: "test"},
			verify: func(t *testing.T, ce *models.CanonicalEvent) {
				if ce.Patient.ID != "PAT123" {
					t.Fatalf("expected PAT123, got %s", ce.Patient.ID)
				}
				if ce.Encounter.Type != "I" {
					t.Fatalf("expected I, got %s", ce.Encounter.Type)
				}
				if ce.Observation.Code != "ORU^R01" {
					t.Fatalf("expected ORU^R01, got %s", ce.Observation.Code)
				}
			},
		},
		{
			name: "Unsupported format",
			ne: &models.NormalizedEvent{
				Format: config.FormatX12,
			},
			env:       api.Envelope{},
			expectErr: true,
		},
	}

	xfm := NewHL7Transformer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// ⭐ UPDATED: Transform now requires ctx
			ce, err := xfm.Transform(context.Background(), tt.ne, tt.env)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tt.verify(t, ce)
		})
	}
}
