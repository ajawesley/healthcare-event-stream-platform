package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

type FakeRouter struct {
	calledPayload []byte
	calledEnv     api.Envelope
	resp          *models.CanonicalEvent
	err           error
}

func (f *FakeRouter) Route(raw []byte, env api.Envelope) (*models.CanonicalEvent, error) {
	f.calledPayload = raw
	f.calledEnv = env
	return f.resp, f.err
}

func TestHandler_RoutesAndReturns202(t *testing.T) {

	fake := &FakeRouter{
		resp: &models.CanonicalEvent{
			Format: "generic",
		},
	}

	h := NewHandler(
		WithRouter(fake),
	)

	body := map[string]any{
		"envelope": mustJSONMarshal(api.Envelope{
			EventID:      "evt-1",
			EventType:    "ingest.test",
			ProducedAt:   time.Now().UTC(),
			SourceSystem: "test",
		}),
		"payload": json.RawMessage(`"x"`),
	}

	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(b))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	// --- Assert status ---
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}

	// --- Assert router was called ---
	if string(fake.calledPayload) != `"x"` {
		t.Fatalf("expected payload \"x\", got %s", string(fake.calledPayload))
	}
	if fake.calledEnv.EventID != "evt-1" {
		t.Fatalf("expected envelope event_id evt-1, got %s", fake.calledEnv.EventID)
	}

	// --- Assert response body ---
	var resp api.IngestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	if resp.EventID != "evt-1" {
		t.Fatalf("expected EventID evt-1, got %s", resp.EventID)
	}
	if resp.Format != "generic" {
		t.Fatalf("expected Format generic, got %s", resp.Format)
	}
	if resp.IngestedAt == "" {
		t.Fatalf("expected IngestedAt to be set")
	}
}
