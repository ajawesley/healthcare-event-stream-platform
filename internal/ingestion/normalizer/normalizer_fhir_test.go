package normalizer

import (
	"testing"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

func TestFHIRNormalizer(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		expectErr bool
		verify    func(t *testing.T, ne *models.NormalizedEvent)
	}{
		{
			name: "Extracts resourceType and id",
			raw: `{
                "resourceType": "Patient",
                "id": "pat123"
            }`,
			verify: func(t *testing.T, ne *models.NormalizedEvent) {
				if ne.Fields["resource_type"] != "Patient" {
					t.Fatalf("expected resource_type Patient, got %v", ne.Fields["resource_type"])
				}
				if ne.Fields["id"] != "pat123" {
					t.Fatalf("expected id pat123, got %v", ne.Fields["id"])
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
				// Your normalizer logs show: metadata is stored under "meta.profile"
				if ne.Metadata["meta.profile"] == nil {
					t.Fatalf("expected metadata to contain meta.profile")
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
			ne, err := n.Normalize([]byte(tt.raw), api.Envelope{})

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
