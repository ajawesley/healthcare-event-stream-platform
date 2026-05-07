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

func TestFHIRTransformer(t *testing.T) {
	xfm := NewFHIRTransformer()

	tests := []struct {
		name      string
		ne        *models.NormalizedEvent
		env       api.Envelope
		expectErr bool
		verify    func(t *testing.T, ce *models.CanonicalEvent)
	}{
		{
			name: "Transforms Patient",
			ne: &models.NormalizedEvent{
				Format: config.FormatFHIR,
				Fields: map[string]any{
					"fhir.resource_type": "Patient",
					"fhir.id":            "pat123",
				},
			},
			env: api.Envelope{EventID: "evt1", SourceSystem: "sys"},
			verify: func(t *testing.T, ce *models.CanonicalEvent) {
				if ce.Patient == nil || ce.Patient.ID != "pat123" {
					t.Fatalf("expected Patient ID pat123, got %+v", ce.Patient)
				}
			},
		},
		{
			name: "Transforms Observation",
			ne: &models.NormalizedEvent{
				Format: config.FormatFHIR,
				Fields: map[string]any{
					"fhir.resource_type": "Observation",
					"fhir.code":          "718-7",
					"fhir.value":         13.5,
				},
			},
			env: api.Envelope{EventID: "evt2", SourceSystem: "sys"},
			verify: func(t *testing.T, ce *models.CanonicalEvent) {
				if ce.Observation == nil {
					t.Fatalf("expected Observation")
				}
				if ce.Observation.Code != "718-7" {
					t.Fatalf("expected code 718-7, got %s", ce.Observation.Code)
				}
				if ce.Observation.Value != 13.5 {
					t.Fatalf("expected value 13.5, got %v", ce.Observation.Value)
				}
			},
		},
		{
			name: "Unsupported format",
			ne: &models.NormalizedEvent{
				Format: config.FormatHL7,
			},
			expectErr: true,
		},
		{
			name: "Missing resourceType",
			ne: &models.NormalizedEvent{
				Format: config.FormatFHIR,
				Fields: map[string]any{},
			},
			expectErr: true,
		},
		{
			name: "Unsupported resourceType",
			ne: &models.NormalizedEvent{
				Format: config.FormatFHIR,
				Fields: map[string]any{
					"fhir.resource_type": "WeirdThing",
				},
			},
			expectErr: true,
		},
	}

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

			if tt.verify != nil {
				tt.verify(t, ce)
			}
		})
	}
}
