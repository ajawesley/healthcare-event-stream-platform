package ingestion

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIngestHandler(t *testing.T) {
	h := NewHandler()

	tests := []struct {
		name           string
		method         string
		body           string
		wantStatusCode int
	}{
		{
			name:   "valid request",
			method: http.MethodPost,
			body: `{
                "envelope": {
                    "event_id": "abc-123",
                    "event_type": "claims.test.created",
                    "produced_at": "2026-05-01T14:23:00Z",
                    "source_system": "unit-test"
                },
                "payload": {"foo":"bar"}
            }`,
			wantStatusCode: http.StatusAccepted,
		},
		{
			name:           "invalid method",
			method:         http.MethodGet,
			body:           "",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:   "missing envelope fields",
			method: http.MethodPost,
			body: `{
                "envelope": {
                    "event_id": "",
                    "event_type": "",
                    "produced_at": "",
                    "source_system": ""
                },
                "payload": {}
            }`,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:           "invalid JSON body",
			method:         http.MethodPost,
			body:           `{ "envelope": "not-an-object" `,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "invalid envelope structure",
			method: http.MethodPost,
			body: `{
                "envelope": 123,
                "payload": {}
            }`,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, "/events/ingest", bytes.NewBufferString(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, "/events/ingest", nil)
			}

			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Fatalf("expected %d, got %d", tt.wantStatusCode, w.Code)
			}
		})
	}
}
