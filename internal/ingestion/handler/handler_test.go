package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/observability"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// -----------------------------------------------------------------------------
// Test Init
// -----------------------------------------------------------------------------

func init() {
	observability.NewLogger("hesp-ecs", "test")
	observability.InitMetrics("hesp-ecs", "test")
	otel.SetTracerProvider(trace.NewNoopTracerProvider())
}

// -----------------------------------------------------------------------------
// Fake Router (updated for ctx-aware interface)
// -----------------------------------------------------------------------------

type FakeRouter struct {
	resp          *models.CanonicalEvent
	err           error
	calledPayload []byte
	calledEnv     api.Envelope
	calledCtx     context.Context
}

func (f *FakeRouter) Route(ctx context.Context, payload []byte, env api.Envelope) (*models.CanonicalEvent, error) {
	f.calledCtx = ctx
	f.calledPayload = payload
	f.calledEnv = env
	return f.resp, f.err
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func mustJSONMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

// -----------------------------------------------------------------------------
// Unified Table-Driven Test Suite
// -----------------------------------------------------------------------------

func TestHandler(t *testing.T) {

	now := time.Now().UTC()

	tests := []struct {
		name           string
		body           map[string]any
		router         *FakeRouter
		expectedStatus int
		verify         func(t *testing.T, fake *FakeRouter, rec *httptest.ResponseRecorder)
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
			router:         &FakeRouter{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid produced_at",
			body: map[string]any{
				"envelope": mustJSONMarshal(api.Envelope{
					EventID:      "evt-1",
					EventType:    "ingest.test",
					ProducedAt:   time.Time{}, // invalid
					SourceSystem: "test",
				}),
				"payload": json.RawMessage(`"x"`),
			},
			router:         &FakeRouter{},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid envelope structure",
			body: map[string]any{
				"envelope": 123, // not JSON object
				"payload":  json.RawMessage(`"x"`),
			},
			router:         &FakeRouter{},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "router failure",
			body: map[string]any{
				"envelope": mustJSONMarshal(api.Envelope{
					EventID:      "evt-1",
					EventType:    "ingest.test",
					ProducedAt:   now,
					SourceSystem: "test",
				}),
				"payload": json.RawMessage(`"x"`),
			},
			router:         &FakeRouter{err: errors.New("router failed")},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "normalization path",
			body: map[string]any{
				"envelope": mustJSONMarshal(api.Envelope{
					EventID:      "evt-1",
					EventType:    "ingest.test",
					ProducedAt:   now,
					SourceSystem: "test",
				}),
				"payload": json.RawMessage(`"x"`),
			},
			router: &FakeRouter{
				resp: &models.CanonicalEvent{
					Format: config.FormatGeneric,
				},
			},
			expectedStatus: http.StatusAccepted,
			verify: func(t *testing.T, fake *FakeRouter, rec *httptest.ResponseRecorder) {
				if string(fake.calledPayload) != `"x"` {
					t.Fatalf("expected payload \"x\", got %s", string(fake.calledPayload))
				}
				if fake.calledEnv.EventID != "evt-1" {
					t.Fatalf("expected envelope event_id evt-1, got %s", fake.calledEnv.EventID)
				}
				if fake.calledCtx == nil {
					t.Fatalf("expected ctx to be passed into router")
				}
			},
		},
		{
			name: "transformation path",
			body: map[string]any{
				"envelope": mustJSONMarshal(api.Envelope{
					EventID:      "evt-2",
					EventType:    "ingest.transform",
					ProducedAt:   now,
					SourceSystem: "test-transform",
				}),
				"payload": json.RawMessage(`"y"`),
			},
			router: &FakeRouter{
				resp: &models.CanonicalEvent{
					Format: config.FormatGeneric,
				},
			},
			expectedStatus: http.StatusAccepted,
			verify: func(t *testing.T, fake *FakeRouter, rec *httptest.ResponseRecorder) {
				if string(fake.calledPayload) != `"y"` {
					t.Fatalf("expected payload \"y\", got %s", string(fake.calledPayload))
				}
				if fake.calledEnv.EventID != "evt-2" {
					t.Fatalf("expected envelope event_id evt-2, got %s", fake.calledEnv.EventID)
				}
				if fake.calledCtx == nil {
					t.Fatalf("expected ctx to be passed into router")
				}
			},
		},
		{
			name: "valid request end-to-end",
			body: map[string]any{
				"envelope": mustJSONMarshal(api.Envelope{
					EventID:      "evt-3",
					EventType:    "ingest.test",
					ProducedAt:   now,
					SourceSystem: "test",
				}),
				"payload": json.RawMessage(`"z"`),
			},
			router: &FakeRouter{
				resp: &models.CanonicalEvent{
					EventID:      "evt-3",
					SourceSystem: "test",
					Format:       config.FormatGeneric,
				},
			},
			expectedStatus: http.StatusAccepted,
			verify: func(t *testing.T, fake *FakeRouter, rec *httptest.ResponseRecorder) {
				var resp api.IngestResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("invalid JSON response: %v", err)
				}
				if resp.EventID != "evt-3" {
					t.Fatalf("expected EventID evt-3, got %s", resp.EventID)
				}
				if resp.Format != "generic" {
					t.Fatalf("expected Format generic, got %s", resp.Format)
				}
				if fake.calledCtx == nil {
					t.Fatalf("expected ctx to be passed into router")
				}
			},
		},
	}

	// -------------------------------------------------------------------------
	// Execute table
	// -------------------------------------------------------------------------
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

			if tt.verify != nil {
				tt.verify(t, tt.router, rec)
			}
		})
	}
}
