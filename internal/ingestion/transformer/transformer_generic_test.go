package transformer

import (
	"context"
	"testing"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

func TestGenericTransformer(t *testing.T) {
	tests := []struct {
		name      string
		ne        *models.NormalizedEvent
		env       api.Envelope
		expectErr bool
		verify    func(t *testing.T, ce *models.CanonicalEvent)
	}{
		{
			name: "Maps generic fields",
			ne: &models.NormalizedEvent{
				Format: config.FormatGeneric,
				Fields: map[string]any{
					"a": 1,
					"b": "two",
				},
			},
			env: api.Envelope{EventID: "evt1", SourceSystem: "test"},
			verify: func(t *testing.T, ce *models.CanonicalEvent) {
				if ce.EventID != "evt1" {
					t.Fatalf("expected EventID=evt1, got %s", ce.EventID)
				}
				if ce.SourceSystem != "test" {
					t.Fatalf("expected SourceSystem=test, got %s", ce.SourceSystem)
				}
				if ce.Format != config.FormatGeneric {
					t.Fatalf("expected Format=generic, got %s", ce.Format)
				}

				fields, ok := ce.Metadata["generic_fields"].(map[string]any)
				if !ok {
					t.Fatalf("expected generic_fields map, got %T", ce.Metadata["generic_fields"])
				}

				if fields["a"] != 1 {
					t.Fatalf("expected a=1, got %v", fields["a"])
				}
				if fields["b"] != "two" {
					t.Fatalf("expected b=two, got %v", fields["b"])
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

	xfm := NewGenericTransformer()

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
