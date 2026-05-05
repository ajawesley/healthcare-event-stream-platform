package router

import (
	"testing"

	"github.com/ajawes/hesp/internal/config"
)

func TestTransformationRouter_TableDriven(t *testing.T) {
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
			_, err := r.TransformerFor(tt.format)
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
