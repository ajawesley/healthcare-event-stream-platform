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
	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/ingestion/normalizer"
)

type FakeNormalizer struct {
	called bool
	err    error
}

func (f *FakeNormalizer) Normalize(value any, meta api.Envelope) (*models.CanonicalEvent, error) {
	f.called = true
	return &models.CanonicalEvent{
		EventID:      meta.EventID,
		SourceSystem: meta.SourceSystem,
		Format:       detector.FormatGeneric,
		RawValue:     value,
	}, f.err
}

type FakeNormalizationRouter struct {
	normalizer normalizer.Normalizer
	err        error
}

func (r *FakeNormalizationRouter) NormalizerFor(format detector.Format) (normalizer.Normalizer, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.normalizer, nil
}

func TestHandler_NormalizationPath(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)

	fakeNormalizer := &FakeNormalizer{}
	fakeNormalizationRouter := &FakeNormalizationRouter{normalizer: fakeNormalizer}

	h := NewHandler(
		WithConfig(minimalConfig()),
		WithRouter(&FakeRouter{format: detector.FormatGeneric}),
		WithDetector(&FakeDetector{format: detector.FormatGeneric}),
		WithTransformationRouter(&FakeTransformationRouter{transformer: &FakeTransformer{}}),
		WithNormalizationRouter(fakeNormalizationRouter),
	)

	body := map[string]any{
		"envelope": mustJSONMarshal(api.Envelope{
			EventID:      "evt-1",
			EventType:    "ingest.test",
			ProducedAt:   now,
			SourceSystem: "test",
		}),
		"payload": json.RawMessage(`"x"`),
	}

	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(b))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
	if !fakeNormalizer.called {
		t.Fatalf("expected normalizer to be called")
	}
}
