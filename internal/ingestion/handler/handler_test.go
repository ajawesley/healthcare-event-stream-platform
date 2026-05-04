package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/detector"
	"github.com/ajawes/hesp/internal/ingestion/router"
)

type FakeDetector struct {
	format detector.Format
}

func (f *FakeDetector) Detect(_ []byte) detector.Format {
	return f.format
}

type FakeRouter struct {
	format detector.Format
	err    error
}

func (f *FakeRouter) Route(_ json.RawMessage) (*router.RoutedPayload, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &router.RoutedPayload{Format: f.format}, nil
}

func TestHandler_TableDriven(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)

	tests := []struct {
		name           string
		body           map[string]any
		router         router.Router
		detector       detector.Detector
		expectedStatus int
	}{
		{
			name: "missing payload",
			body: map[string]any{
				"envelope": mustJSONMarshal(api.Envelope{
					EventID:      "evt-1",
					EventType:    "ingest.test",
					ProducedAt:   now,
					SourceSystem: "test",
				}),
			},
			router:         &FakeRouter{format: detector.FormatGeneric},
			detector:       &FakeDetector{format: detector.FormatGeneric},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid produced_at",
			body: map[string]any{
				"envelope": mustJSONMarshal(api.Envelope{
					EventID:      "evt-1",
					EventType:    "ingest.test",
					ProducedAt:   "not-a-timestamp",
					SourceSystem: "test",
				}),
				"payload": json.RawMessage(`"x"`),
			},
			router:         &FakeRouter{format: detector.FormatGeneric},
			detector:       &FakeDetector{format: detector.FormatGeneric},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid envelope structure",
			body: map[string]any{
				"envelope": 123,
				"payload":  json.RawMessage(`"x"`),
			},
			router:         &FakeRouter{format: detector.FormatGeneric},
			detector:       &FakeDetector{format: detector.FormatGeneric},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "valid request",
			body: map[string]any{
				"envelope": mustJSONMarshal(api.Envelope{
					EventID:      "evt-1",
					EventType:    "ingest.test",
					ProducedAt:   now,
					SourceSystem: "test",
				}),
				"payload": json.RawMessage(`"x"`),
			},
			router:         &FakeRouter{format: detector.FormatGeneric},
			detector:       &FakeDetector{format: detector.FormatGeneric},
			expectedStatus: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(
				WithConfig(minimalConfig()),
				WithRouter(tt.router),
				WithDetector(tt.detector),
			)

			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(b))
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func mustJSONMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func minimalConfig() detector.DetectorConfig {
	return detector.DetectorConfig{
		Rules: []detector.DetectionRule{
			{Name: "generic", Format: detector.FormatGeneric},
		},
	}
}
