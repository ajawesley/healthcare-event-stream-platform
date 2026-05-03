package ingestion

import "testing"

func TestDetector_Detect(t *testing.T) {
	cfg := DetectorConfig{
		Rules: []DetectionRule{
			{Name: "hl7", Format: FormatHL7, Prefix: "MSH|"},
			{Name: "x12", Format: FormatX12, Prefix: "ISA*"},
			{Name: "fhir", Format: FormatFHIR, ContainsKey: "resourceType"},
		},
	}
	d := NewDetector(cfg)

	tests := []struct {
		name   string
		input  []byte
		format Format
	}{
		{"hl7", []byte("MSH|^~\\&|"), FormatHL7},
		{"x12", []byte("ISA*00*          *00*          *"), FormatX12},
		{"fhir", []byte(`{"resourceType":"Patient"}`), FormatFHIR},
		{"generic", []byte(`{"foo":"bar"}`), FormatGeneric},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.Detect(tt.input)
			if got != tt.format {
				t.Fatalf("expected %s, got %s", tt.format, got)
			}
		})
	}
}
