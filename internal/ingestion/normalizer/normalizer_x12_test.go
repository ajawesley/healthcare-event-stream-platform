package normalizer

import (
	"testing"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

func TestX12Normalizer_Normalize(t *testing.T) {
	raw := `ISA*00*          *00*          *ZZ*SENDERID       *ZZ*RECEIVERID     *210101*1200*U*00401*000000001*0*T*:~
GS*HC*SENDER*RECEIVER*20250101*1200*1*X*005010X222A1~
ST*837*0001~
NM1*IL*1*Doe*John****MI*PAT123~
CLM*ENC456*100***11:B:1*Y*A*Y*I~
DTP*472*D8*20250101~`

	env := api.Envelope{
		EventID:      "evt-1",
		SourceSystem: "test",
	}

	n := NewX12Normalizer()

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
	if ce.Observation.Code != "837" {
		t.Fatalf("expected observation code 837, got %s", ce.Observation.Code)
	}
}

func TestX12Normalizer_MissingSegments(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want error
	}{
		{"missing ISA", "GS*HC*X*Y~ST*837*1~", models.ErrX12MissingISA},
		{"missing GS", "ISA*00*00~ST*837*1~", models.ErrX12MissingGS},
		{"missing ST", "ISA*00*00~GS*HC*X*Y~", models.ErrX12MissingST},
		{"missing NM1 IL", "ISA*00*00~GS*HC*X*Y~ST*837*1~CLM*1~", models.ErrX12MissingNM1IL},
		{"missing CLM", "ISA*00*00~GS*HC*X*Y~ST*837*1~NM1*IL*1*Doe*John~", models.ErrX12MissingCLM},
	}

	n := NewX12Normalizer()
	env := api.Envelope{EventID: "evt", SourceSystem: "test"}

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
