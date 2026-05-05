package transformer

import (
	"testing"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

func TestX12Transformer(t *testing.T) {
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
				Format: config.FormatX12,
				Fields: map[string]any{
					"nm1.il.patient_id":   "PAT123",
					"nm1.il.first_name":   "John",
					"nm1.il.last_name":    "Doe",
					"clm.encounter_id":    "ENC456",
					"st.transaction_code": "837",
				},
			},
			env: api.Envelope{EventID: "evt1", SourceSystem: "test"},
			verify: func(t *testing.T, ce *models.CanonicalEvent) {
				if ce.Patient.ID != "PAT123" {
					t.Fatalf("expected PAT123, got %s", ce.Patient.ID)
				}
				if ce.Encounter.ID != "ENC456" {
					t.Fatalf("expected ENC456, got %s", ce.Encounter.ID)
				}
				if ce.Observation.Code != "837" {
					t.Fatalf("expected 837, got %s", ce.Observation.Code)
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
	}

	xfm := NewX12Transformer()

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
