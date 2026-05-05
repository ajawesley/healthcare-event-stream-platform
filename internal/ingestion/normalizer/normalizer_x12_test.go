package normalizer

import (
	"testing"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

func TestX12Normalizer(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		expectErr error
		verify    func(t *testing.T, ne *models.NormalizedEvent)
	}{
		{
			name: "Basic extraction",
			raw: `ISA*00*          *00*          *ZZ*SENDERID       *ZZ*RECEIVERID     *210101*1200*U*00401*000000001*0*T*:~
GS*HC*SENDER*RECEIVER*20250101*1200*1*X*005010X222A1~
ST*837*0001~
NM1*IL*1*Doe*John****MI*PAT123~
CLM*ENC456*100***11:B:1*Y*A*Y*I~`,
			verify: func(t *testing.T, ne *models.NormalizedEvent) {

				if ne.Fields["nm1.first_name"] != "John" {
					t.Fatalf("expected first name John, got %v", ne.Fields["nm1.first_name"])
				}

				if ne.Fields["nm1.last_name"] != "Doe" {
					t.Fatalf("expected last name Doe, got %v", ne.Fields["nm1.last_name"])
				}

				if ne.Fields["nm1.patient_id"] != "PAT123" {
					t.Fatalf("expected patient id PAT123, got %v", ne.Fields["nm1.patient_id"])
				}

				if ne.Fields["clm.encounter_id"] != "ENC456" {
					t.Fatalf("expected encounter id ENC456, got %v", ne.Fields["clm.encounter_id"])
				}

				if ne.Fields["st.transaction_set"] != "837" {
					t.Fatalf("expected transaction code 837, got %v", ne.Fields["st.transaction_set"])
				}
			},
		},
		{"Missing ISA", "GS*HC*X*Y~ST*837*1~", models.ErrX12MissingISA, nil},
		{"Missing GS", "ISA*00*00~ST*837*1~", models.ErrX12MissingGS, nil},
		{"Missing ST", "ISA*00*00~GS*HC*X*Y~", models.ErrX12MissingST, nil},
		{"Missing NM1 IL", "ISA*00*00~GS*HC*X*Y~ST*837*1~CLM*1~", models.ErrX12MissingNM1IL, nil},
		{"Missing CLM", "ISA*00*00~GS*HC*X*Y~ST*837*1~NM1*IL*1*Doe*John~", models.ErrX12MissingCLM, nil},
	}

	n := NewX12Normalizer()
	env := api.Envelope{EventID: "evt", SourceSystem: "test"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ne, err := n.Normalize([]byte(tt.raw), env)

			if tt.expectErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectErr)
				}
				if err != tt.expectErr {
					t.Fatalf("expected %v, got %v", tt.expectErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.verify != nil {
				tt.verify(t, ne)
			}
		})
	}
}
