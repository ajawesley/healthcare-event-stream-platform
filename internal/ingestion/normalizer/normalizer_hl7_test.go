package normalizer

import (
	"context"
	"testing"

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

func TestHL7Normalizer(t *testing.T) {
	tests := []struct {
		name      string
		raw       []byte
		wantErr   error
		wantCheck func(t *testing.T, ce *models.NormalizedEvent)
	}{
		{
			name: "success",
			raw: []byte(`MSH|^~\&|LAB|HOSP|EHR|HOSP|202501011200||ORU^R01|123|P|2.3
PID|1||PAT123||Doe^John
PV1|1|I|WARD^101|||||||||||||||||ENC456`),
			wantErr: nil,
			wantCheck: func(t *testing.T, ce *models.NormalizedEvent) {
				if ce.Fields["pid.id"] != "PAT123" {
					t.Fatalf("expected patient ID PAT123, got %s", ce.Fields["pid.id"])
				}
				if ce.Fields["pid.first_name"] != "John" {
					t.Fatalf("expected first name John, got %s", ce.Fields["pid.first_name"])
				}
				if ce.Fields["pid.last_name"] != "Doe" {
					t.Fatalf("expected last name Doe, got %s", ce.Fields["pid.last_name"])
				}
				if ce.Fields["pv1.encounter_id"] != "ENC456" {
					t.Fatalf("expected encounter ID ENC456, got %s", ce.Fields["pv1.encounter_id"])
				}
				if ce.Fields["msh.message_type"] != "ORU^R01" {
					t.Fatalf("expected message type ORU^R01, got %s", ce.Fields["msh.message_type"])
				}
			},
		},
		{"missing MSH", []byte("PID|1||PAT123\nPV1|1|I"), models.ErrHL7MissingMSH, nil},
		{"missing PID", []byte("MSH|a|b\nPV1|1|I"), models.ErrHL7MissingPID, nil},
		{"missing PV1", []byte("MSH|a|b\nPID|1||PAT123"), models.ErrHL7MissingPV1, nil},
	}

	n := NewHL7Normalizer()
	env := api.Envelope{EventID: "evt-1", SourceSystem: "test"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// ⭐ UPDATED: Normalize now requires ctx
			ce, err := n.Normalize(context.Background(), tt.raw, env)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if err != tt.wantErr {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantCheck != nil {
				tt.wantCheck(t, ce)
			}
		})
	}
}
