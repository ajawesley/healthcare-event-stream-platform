package detector

import (
	"testing"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/observability"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	observability.NewLogger("hesp-ecs", "test")
	observability.InitMetrics("hesp-ecs", "test")
	otel.SetTracerProvider(trace.NewNoopTracerProvider())
}

func TestDetect(t *testing.T) {

	tests := []struct {
		name     string
		rules    []config.Rule
		payload  []byte
		expected config.Format
	}{
		{
			name: "prefix match",
			rules: []config.Rule{
				{Name: "hl7", Prefix: "MSH", Format: config.FormatHL7},
			},
			payload:  []byte("MSH|^~\\&|..."),
			expected: config.FormatHL7,
		},
		{
			name: "json key match",
			rules: []config.Rule{
				{Name: "fhir", ContainsKey: "resourceType", Format: config.FormatFHIR},
			},
			payload:  []byte(`{"resourceType":"Patient"}`),
			expected: config.FormatFHIR,
		},
		{
			name: "nested json key match",
			rules: []config.Rule{
				{Name: "fhir", ContainsKey: "id", Format: config.FormatFHIR},
			},
			payload:  []byte(`{"a":{"b":{"id":"123"}}}`),
			expected: config.FormatFHIR,
		},
		{
			name: "invalid json falls back",
			rules: []config.Rule{
				{Name: "fhir", ContainsKey: "resourceType", Format: config.FormatFHIR},
			},
			payload:  []byte(`{invalid json`),
			expected: config.FormatGeneric,
		},
		{
			name: "no rules match → fallback",
			rules: []config.Rule{
				{Name: "hl7", Prefix: "MSH", Format: config.FormatHL7},
			},
			payload:  []byte("XYZ123"),
			expected: config.FormatGeneric,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Inject rules directly
			d := &detectorImpl{rules: tt.rules}

			got := d.Detect(tt.payload)
			if got != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}
