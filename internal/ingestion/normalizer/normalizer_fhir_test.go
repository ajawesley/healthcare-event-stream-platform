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

func TestFHIRNormalizer(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		expectErr bool
		verify    func(t *testing.T, ne *models.NormalizedEvent)
	}{
		{
			name: "Extracts resourceType and id (namespaced)",
			raw: `{
                "resourceType": "Patient",
                "id": "pat123"
            }`,
			verify: func(t *testing.T, ne *models.NormalizedEvent) {
				if ne.Fields["fhir.resource_type"] != "Patient" {
					t.Fatalf("expected fhir.resource_type Patient, got %v", ne.Fields["fhir.resource_type"])
				}
				if ne.Fields["fhir.id"] != "pat123" {
					t.Fatalf("expected fhir.id pat123, got %v", ne.Fields["fhir.id"])
				}
			},
		},
		{
			name: "Extracts metadata",
			raw: `{
                "resourceType": "Patient",
                "id": "pat123",
                "meta": { "profile": ["x"] }
            }`,
			verify: func(t *testing.T, ne *models.NormalizedEvent) {
				if ne.Metadata["meta.profile"] == nil {
					t.Fatalf("expected metadata to contain meta.profile")
				}
			},
		},
		{
			name: "Extracts Observation code and value",
			raw: `{
                "resourceType": "Observation",
                "id": "obs1",
                "code": {
                    "coding": [{
                        "system": "http://loinc.org",
                        "code": "718-7"
                    }]
                },
                "valueQuantity": {
                    "value": 13.5
                }
            }`,
			verify: func(t *testing.T, ne *models.NormalizedEvent) {
				if ne.Fields["fhir.resource_type"] != "Observation" {
					t.Fatalf("expected fhir.resource_type Observation, got %v", ne.Fields["fhir.resource_type"])
				}
				if ne.Fields["fhir.id"] != "obs1" {
					t.Fatalf("expected fhir.id obs1, got %v", ne.Fields["fhir.id"])
				}
				if ne.Fields["fhir.code"] != "718-7" {
					t.Fatalf("expected fhir.code 718-7, got %v", ne.Fields["fhir.code"])
				}
				if ne.Fields["fhir.value"] != 13.5 {
					t.Fatalf("expected fhir.value 13.5, got %v", ne.Fields["fhir.value"])
				}
			},
		},
		{
			name:      "Invalid JSON returns error",
			raw:       `{ invalid json }`,
			expectErr: true,
		},
	}

	n := NewFHIRNormalizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// ⭐ UPDATED: Normalize now requires ctx
			ne, err := n.Normalize(context.Background(), []byte(tt.raw), api.Envelope{})

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.verify != nil {
				tt.verify(t, ne)
			}
		})
	}
}
