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

func TestParseHL7(t *testing.T) {

	tests := []struct {
		name      string
		raw       string
		expectErr error
		verify    func(t *testing.T, msg *HL7Message)
	}{
		{
			name: "valid HL7 message",
			raw: `MSH|^~\&|LAB|HOSP|EHR|HOSP|202501011200||ORU^R01|123|P|2.3
PID|1||PAT123||Doe^John
PV1|1|I|WARD^101|||||||||||||||||ENC456`,
			expectErr: nil,
			verify: func(t *testing.T, msg *HL7Message) {
				if msg.MSH[0] != "MSH" {
					t.Fatalf("expected MSH segment, got %v", msg.MSH)
				}
				if msg.PID[0] != "PID" {
					t.Fatalf("expected PID segment, got %v", msg.PID)
				}
				if msg.PV1[0] != "PV1" {
					t.Fatalf("expected PV1 segment, got %v", msg.PV1)
				}
			},
		},
		{
			name:      "missing MSH",
			raw:       "PID|1||PAT123\nPV1|1|I",
			expectErr: ErrHL7MissingMSH,
		},
		{
			name:      "missing PID",
			raw:       "MSH|a|b\nPV1|1|I",
			expectErr: ErrHL7MissingPID,
		},
		{
			name:      "missing PV1",
			raw:       "MSH|a|b\nPID|1||PAT123",
			expectErr: ErrHL7MissingPV1,
		},
		{
			name: "handles CRLF",
			raw:  "MSH|a|b\r\nPID|1||PAT123\r\nPV1|1|I",
			verify: func(t *testing.T, msg *HL7Message) {
				if msg.PID[2] != "" {
					t.Fatalf("expected PID field 2 empty, got %s", msg.PID[2])
				}
			},
		},
		{
			name: "single-line HL7",
			raw:  "MSH|a|b|c|d|e|f|g|ORU^R01|123|P|2.3\nPID|1||PAT123\nPV1|1|I",
			verify: func(t *testing.T, msg *HL7Message) {
				if msg.MSH == nil || msg.PID == nil || msg.PV1 == nil {
					t.Fatalf("expected all segments parsed")
				}
			},
		},
		{
			name:      "empty payload",
			raw:       "",
			expectErr: ErrHL7MissingMSH,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ParseHL7(tt.raw)

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
