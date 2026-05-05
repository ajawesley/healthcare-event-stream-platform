package transformer

import (
	"testing"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

func TestFHIRTransformer(t *testing.T) {
	tests := []struct {
		name      string
		ne        *models.NormalizedEvent
		env       api.Envelope
		expectErr bool
		verify    func(t *testing.T, ce *models.CanonicalEvent)
	}{
		{
			name: "Patient mapping",
			ne: &models.NormalizedEvent{
				Format: config.FormatFHIR,
				Fields: map[string]any{
					"fhir.resource_type": "Patient",
					"fhir.id":            "pat123",
				},
			},
			env: api.Envelope{EventID: "evt1", SourceSystem: "test"},
			verify: func(t *testing.T, ce *models.CanonicalEvent) {
				if ce.Patient.ID != "pat123" {
					t.Fatalf("expected pat123, got %s", ce.Patient.ID)
				}
			},
		},
		{
			name: "Encounter mapping",
			ne: &models.NormalizedEvent{
				Format: config.FormatFHIR,
				Fields: map[string]any{
					"fhir.resource_type": "Encounter",
					"fhir.id":            "enc789",
					"fhir.class.code":    "AMB",
				},
			},
			env: api.Envelope{EventID: "evt2", SourceSystem: "test"},
			verify: func(t *testing.T, ce *models.CanonicalEvent) {
				if ce.Encounter.ID != "enc789" {
					t.Fatalf("expected enc789, got %s", ce.Encounter.ID)
				}
				if ce.Encounter.Type != "AMB" {
					t.Fatalf("expected AMB, got %s", ce.Encounter.Type)
				}
			},
		},
		{
			name: "Observation mapping",
			ne: &models.NormalizedEvent{
				Format: config.FormatFHIR,
				Fields: map[string]any{
					"fhir.resource_type": "Observation",
					"fhir.code":          "12345-6",
					"fhir.value":         "98.6",
				},
			},
			env: api.Envelope{EventID: "evt3", SourceSystem: "test"},
			verify: func(t *testing.T, ce *models.CanonicalEvent) {
				if ce.Observation.Code != "12345-6" {
					t.Fatalf("expected 12345-6, got %s", ce.Observation.Code)
				}
				if ce.Observation.Value != "98.6" {
					t.Fatalf("expected 98.6, got %v", ce.Observation.Value)
				}
			},
		},
		{
			name: "Unsupported resourceType",
			ne: &models.NormalizedEvent{
				Format: config.FormatFHIR,
				Fields: map[string]any{
					"fhir.resource_type": "Medication",
				},
			},
			expectErr: true,
		},
	}

	xfm := NewFHIRTransformer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce, err := xfm.Transform(tt.ne, tt.env)

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
