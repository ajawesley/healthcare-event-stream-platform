package router

import (
	"encoding/json"
	"testing"

	"github.com/ajawes/hesp/internal/ingestion/detector"
)

type fakeDetector struct {
	format detector.Format
	err    error
}

func (f fakeDetector) Detect(raw []byte) detector.Format {
	return f.format
}

func TestParser(t *testing.T) {
	tests := []struct {
		name   string
		format detector.Format
		msg    json.RawMessage
		error  bool
	}{
		{"HL7", detector.FormatHL7, json.RawMessage(`"MSH|^~\\&|LAB|HOSP"`), false},
		{"X12", detector.FormatX12, json.RawMessage(`"ISA*00*00*ZZ*SENDER"`), false},
		{"FHIR", detector.FormatFHIR, json.RawMessage(`{"resourceType":"Patient","id":"123"}`), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewParserRouter(fakeDetector{format: tt.format})

			out, err := r.Route(tt.msg)
			if !tt.error && (err != nil) {
				t.Fatalf("unexpected error: %v", err)
			}

			if out.Format != tt.format {
				t.Fatalf("expected %s, got %s", tt.format, out.Format)
			}
		})
	}
}
