package models

import (
	"testing"

	"github.com/ajawes/hesp/internal/observability"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	observability.NewLogger("hesp-ecs", "test")
	observability.InitMetrics("hesp-ecs", "test")
	otel.SetTracerProvider(trace.NewNoopTracerProvider())
}

func TestParseX12(t *testing.T) {

	tests := []struct {
		name      string
		raw       string
		expectErr error
		verify    func(t *testing.T, msg *X12Message)
	}{
		{
			name: "valid X12 message",
			raw: `ISA*00*          *00*          *ZZ*SENDER*ZZ*RECEIVER*210101*1200*U*00401*000000001*0*T*:~
GS*HC*SENDER*RECEIVER*20250101*1200*1*X*005010X222A1~
ST*837*0001~
NM1*IL*1*Doe*John****MI*PAT123~
CLM*ENC456*100***11:B:1*Y*A*Y*I~
DTP*472*D8*20250101~`,
			verify: func(t *testing.T, msg *X12Message) {
				if msg.ISA == nil {
					t.Fatalf("expected ISA segment")
				}
				if msg.GS == nil {
					t.Fatalf("expected GS segment")
				}
				if msg.ST == nil {
					t.Fatalf("expected ST segment")
				}
				if msg.NM1Patient == nil {
					t.Fatalf("expected NM1*IL segment")
				}
				if msg.CLM == nil {
					t.Fatalf("expected CLM segment")
				}
			},
		},
		{
			name:      "missing ISA",
			raw:       "GS*HC*X*Y~ST*837*1~NM1*IL*1*Doe*John~CLM*1~",
			expectErr: ErrX12MissingISA,
		},
		{
			name:      "missing GS",
			raw:       "ISA*00*00~ST*837*1~NM1*IL*1*Doe*John~CLM*1~",
			expectErr: ErrX12MissingGS,
		},
		{
			name:      "missing ST",
			raw:       "ISA*00*00~GS*HC*X*Y~NM1*IL*1*Doe*John~CLM*1~",
			expectErr: ErrX12MissingST,
		},
		{
			name:      "missing NM1 IL",
			raw:       "ISA*00*00~GS*HC*X*Y~ST*837*1~CLM*1~",
			expectErr: ErrX12MissingNM1IL,
		},
		{
			name:      "missing CLM",
			raw:       "ISA*00*00~GS*HC*X*Y~ST*837*1~NM1*IL*1*Doe*John~",
			expectErr: ErrX12MissingCLM,
		},
		{
			name: "handles whitespace",
			raw: `
ISA*00*00~
GS*HC*X*Y~
ST*837*1~
NM1*IL*1*Doe*John~
CLM*1~
`,
			verify: func(t *testing.T, msg *X12Message) {
				if msg.CLM[0] != "CLM" {
					t.Fatalf("expected CLM segment")
				}
			},
		},
		{
			name:      "empty payload",
			raw:       "",
			expectErr: ErrX12MissingISA,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ParseX12(tt.raw)

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
				tt.verify(t, msg)
			}
		})
	}
}
