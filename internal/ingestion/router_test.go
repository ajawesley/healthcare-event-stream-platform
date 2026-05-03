package ingestion

import "testing"

func TestRouter_RoutesByFormat(t *testing.T) {
	cfg := DetectorConfig{
		Rules: []DetectionRule{
			{
				Name:   "hl7_msh_prefix",
				Format: FormatHL7,
				Prefix: "MSH|",
			},
			{
				Name:   "x12_isa_prefix",
				Format: FormatX12,
				Prefix: "ISA*",
			},
			{
				Name:        "fhir_resource_type",
				Format:      FormatFHIR,
				ContainsKey: "resourceType",
			},
		},
	}

	d := NewDetector(cfg)
	r := NewRouter(d)

	tests := []struct {
		name     string
		payload  []byte
		expected Format
	}{
		{"hl7", []byte("MSH|^~\\&"), FormatHL7},
		{"x12", []byte("ISA*00*"), FormatX12},
		{"fhir", []byte(`{"resourceType":"Patient"}`), FormatFHIR},
		{"generic", []byte("just some text"), FormatGeneric},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			routed, err := r.Route(tt.payload)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if routed.Format != tt.expected {
				t.Fatalf("expected format %q, got %q", tt.expected, routed.Format)
			}
		})
	}
}
