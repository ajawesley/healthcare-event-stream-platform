package ingestion

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type FakeTransformer struct {
	called bool
	err    error
}

func (f *FakeTransformer) Transform(value any) (any, error) {
	f.called = true
	return value, f.err
}

type FakeTransformationRouter struct {
	transformer Transformer
	err         error
}

func (r *FakeTransformationRouter) TransformerFor(format Format) (Transformer, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.transformer, nil
}

func TestHandler_TransformationPath(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)

	fakeTransformer := &FakeTransformer{}
	fakeRouter := &FakeTransformationRouter{transformer: fakeTransformer}

	h := NewHandler(
		WithConfig(minimalConfig()),
		WithRouter(&FakeRouter{format: FormatGeneric}),
		WithDetector(&FakeDetector{format: FormatGeneric}),
		WithTransformationRouter(fakeRouter),
	)

	body := map[string]any{
		"envelope": mustJSONMarshal(envelope{
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
	if !fakeTransformer.called {
		t.Fatalf("expected transformer to be called")
	}
}
