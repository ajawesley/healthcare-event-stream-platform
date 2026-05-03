package ingestion

import (
	"encoding/json"
	"testing"
)

type fakeDetector struct {
	format Format
	err    error
}

func (f fakeDetector) Detect(raw []byte) Format {
	return f.format
}

func TestParserRouter_RouteHL7(t *testing.T) {
	raw := json.RawMessage(`"MSH|^~\\&|LAB|HOSP"`)

	r := NewParserRouter(fakeDetector{format: FormatHL7})

	out, err := r.Route(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Format != FormatHL7 {
		t.Fatalf("expected %s, got %s", FormatHL7, out.Format)
	}
}

func TestParserRouter_RouteX12(t *testing.T) {
	raw := json.RawMessage(`"ISA*00*00*ZZ*SENDER"`)

	r := NewParserRouter(fakeDetector{format: FormatX12})

	out, err := r.Route(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Format != FormatX12 {
		t.Fatalf("expected %s, got %s", FormatX12, out.Format)
	}
}

func TestParserRouter_Unsupported(t *testing.T) {
	raw := json.RawMessage(`"???bad"`)

	r := NewParserRouter(fakeDetector{format: "unknown"})

	_, err := r.Route(raw)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
