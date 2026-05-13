package api

import (
	"testing"
	"time"

	"github.com/ajawes/hesp/internal/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	observability.NewLogger("hesp-ecs", "test")
	observability.InitMetrics("hesp-ecs", "test")
	otel.SetTracerProvider(trace.NewNoopTracerProvider())
}

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		raw       []byte
		expectErr bool
		verify    func(t *testing.T, e Envelope)
	}{
		{
			name: "valid envelope",
			raw: []byte(`{
                "event_id": "abc123",
                "event_type": "test_event",
                "source_system": "unit_test",
                "produced_at": "2025-01-01T12:00:00Z"
            }`),
			expectErr: false,
			verify: func(t *testing.T, e Envelope) {
				if e.EventID != "abc123" {
					t.Fatalf("expected event_id abc123, got %s", e.EventID)
				}
				if e.EventType != "test_event" {
					t.Fatalf("expected event_type test_event, got %s", e.EventType)
				}
				if e.SourceSystem != "unit_test" {
					t.Fatalf("expected source_system unit_test, got %s", e.SourceSystem)
				}
				if e.ProducedAt != time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC) {
					t.Fatalf("unexpected produced_at: %v", e.ProducedAt)
				}
			},
		},
		{
			name: "invalid timestamp",
			raw: []byte(`{
                "event_id": "abc123",
                "event_type": "test_event",
                "source_system": "unit_test",
                "produced_at": "not-a-timestamp"
            }`),
			expectErr: true,
		},
		{
			name:      "invalid json",
			raw:       []byte(`{invalid json`),
			expectErr: true,
		},
		{
			name: "missing produced_at",
			raw: []byte(`{
                "event_id": "abc123",
                "event_type": "test_event",
                "source_system": "unit_test"
            }`),
			expectErr: true,
		},
		{
			name:      "empty object",
			raw:       []byte(`{}`),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var env Envelope
			err := env.UnmarshalJSON(tt.raw)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.verify != nil {
				tt.verify(t, env)
			}
		})
	}
}
