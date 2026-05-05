package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

type CanonicalFakeRouter struct {
	Canonical *models.CanonicalEvent
	Err       error
}

func (f *CanonicalFakeRouter) Route(raw []byte, env api.Envelope) (*models.CanonicalEvent, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Canonical, nil
}

func TestHandler_TableDriven(t *testing.T) {

	tests := []struct {
		name           string
		body           map[string]any
		router         *CanonicalFakeRouter
		expectedStatus int
	}{
		{
			name: "missing payload",
			body: map[string]any{
				"envelope": mustJSONMarshal(api.Envelope{
					EventID:      "evt-1",
					EventType:    "ingest.test",
					ProducedAt:   time.Now().UTC(),
					SourceSystem: "test",
				}),
			},
			router:         &CanonicalFakeRouter{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid produced_at",
			body: map[string]any{
				"envelope": mustJSONMarshal(api.Envelope{
					EventID:      "evt-1",
					EventType:    "ingest.test",
					ProducedAt:   time.Time{}, // invalid time
					SourceSystem: "test",
				}),
				"payload": json.RawMessage(`"x"`),
			},
			router:         &CanonicalFakeRouter{},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid envelope structure",
			body: map[string]any{
				"envelope": 123,
				"payload":  json.RawMessage(`"x"`),
			},
			router:         &CanonicalFakeRouter{},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "router failure",
			body: map[string]any{
				"envelope": mustJSONMarshal(api.Envelope{
					EventID:      "evt-1",
					EventType:    "ingest.test",
					ProducedAt:   time.Time{}, // invalid time to trigger envelope validation failure
					SourceSystem: "test",
				}),
				"payload": json.RawMessage(`"x"`),
			},
			router:         &CanonicalFakeRouter{Err: errors.New("router failed")},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "valid request",
			body: map[string]any{
				"envelope": mustJSONMarshal(api.Envelope{
					EventID:      "evt-1",
					EventType:    "ingest.test",
					ProducedAt:   time.Now().UTC(),
					SourceSystem: "test",
				}),
				"payload": json.RawMessage(`"x"`),
			},
			router: &CanonicalFakeRouter{
				Canonical: &models.CanonicalEvent{
					EventID:      "evt-1",
					SourceSystem: "test",
					Format:       config.FormatGeneric,
				},
			},
			expectedStatus: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(
				WithRouter(tt.router),
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
