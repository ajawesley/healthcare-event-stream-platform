package ingestion

import "testing"

func TestTransformationRouter_TableDriven(t *testing.T) {
	r := NewTransformationRouter()

	tests := []struct {
		name      string
		format    Format
		expectErr bool
	}{
		{"hl7", FormatHL7, false},
		{"x12", FormatX12, false},
		{"fhir", FormatFHIR, false},
		{"generic", FormatGeneric, false},
		{"unknown", Format("weird"), true},
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
