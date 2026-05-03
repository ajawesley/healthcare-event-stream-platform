package ingestion

import (
	"testing"
)

func TestHL7Normalizer_Normalize(t *testing.T) {
	raw := `MSH|^~\&|LAB|HOSP|EHR|HOSP|202501011200||ORU^R01|123|P|2.3
PID|1||PAT123||Doe^John
PV1|1|I|WARD^101|||||||||||||||||ENC456`

	env := envelope{
		EventID:      "evt-1",
		SourceSystem: "test",
	}

	n := NewHL7Normalizer()

	ce, err := n.Normalize(raw, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ce.Patient.ID != "PAT123" {
		t.Fatalf("expected patient ID PAT123, got %s", ce.Patient.ID)
	}
	if ce.Patient.FirstName != "John" {
		t.Fatalf("expected first name John, got %s", ce.Patient.FirstName)
	}
	if ce.Patient.LastName != "Doe" {
		t.Fatalf("expected last name Doe, got %s", ce.Patient.LastName)
	}
	if ce.Encounter.ID != "ENC456" {
		t.Fatalf("expected encounter ID ENC456, got %s", ce.Encounter.ID)
	}
	if ce.Observation.Code != "ORU^R01" {
		t.Fatalf("expected observation code ORU^R01, got %s", ce.Observation.Code)
	}
}

func TestHL7Normalizer_MissingSegments(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want error
	}{
		{"missing MSH", "PID|1||PAT123\nPV1|1|I", ErrHL7MissingMSH},
		{"missing PID", "MSH|a|b\nPV1|1|I", ErrHL7MissingPID},
		{"missing PV1", "MSH|a|b\nPID|1||PAT123", ErrHL7MissingPV1},
	}

	n := NewHL7Normalizer()
	env := envelope{EventID: "evt", SourceSystem: "test"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := n.Normalize(tt.raw, env)
			if err == nil {
				t.Fatalf("expected error %v, got nil", tt.want)
			}
			if err != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
		})
	}
}
