package transformer

import (
	"testing"

	"github.com/ajawes/hesp/internal/ingestion/detector"
)

func TestTransformationRouter_TableDriven(t *testing.T) {
	r := NewTransformationRouter()

	tests := []struct {
		name      string
		format    detector.Format
		expectErr bool
	}{
		{"hl7", detector.FormatHL7, false},
		{"x12", detector.FormatX12, false},
		{"fhir", detector.FormatFHIR, false},
		{"generic", detector.FormatGeneric, false},
		{"unknown", detector.Format("weird"), true},
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
