package ingestion

import "testing"

func TestNormalizationRouter_TableDriven(t *testing.T) {
	r := NewNormalizationRouter()

	tests := []struct {
		name      string
		format    Format
		expectErr bool
	}{
		{"hl7", FormatHL7, false},
		{"x12", FormatX12, false},
		{"fhir", FormatFHIR, false},
		{"generic", FormatGeneric, false},
		{"unknown", Format("nope"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := r.NormalizerFor(tt.format)
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
