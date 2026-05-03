package ingestion

import "testing"

func TestParseFHIR(t *testing.T) {
	tests := []struct {
		name      string
		raw       []byte
		expectErr bool
	}{
		{
			name:      "valid FHIR JSON",
			raw:       []byte(`{"resourceType":"Patient"}`),
			expectErr: false,
		},
		{
			name:      "invalid JSON",
			raw:       []byte(`{invalid json`),
			expectErr: true,
		},
		{
			name:      "empty payload",
			raw:       []byte{},
			expectErr: true,
		},
		{
			name:      "non-object JSON",
			raw:       []byte(`"string"`),
			expectErr: true, // cannot unmarshal into map[string]any / FHIR resource
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseFHIR(tt.raw)
			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
