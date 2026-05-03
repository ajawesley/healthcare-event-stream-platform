package ingestion

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestIngestHandler(t *testing.T) {
	// Ensure tests do not depend on external config
	os.Unsetenv("INGESTION_DETECTION_CONFIG")

	h := NewHandler()

	tests := []struct {
		name           string
		method         string
		body           string
		wantStatusCode int
	}{
		// ------------------------------
		// VALID REQUEST
		// ------------------------------
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

		// ------------------------------
		// VALID REQUEST WITH PRIMITIVE PAYLOAD
		// ------------------------------
		{
			name:   "primitive payload",
			method: http.MethodPost,
			body: `{
                "envelope": {
                    "event_id": "abc",
                    "event_type": "claims.test.created",
                    "produced_at": "2026-05-01T14:23:00Z",
                    "source_system": "unit-test"
                },
                "payload": 12345
            }`,
			wantStatusCode: http.StatusAccepted,
		},

		// ------------------------------
		// INVALID METHOD
		// ------------------------------
		{
			name:           "invalid method",
			method:         http.MethodGet,
			body:           "",
			wantStatusCode: http.StatusMethodNotAllowed,
		},

		// ------------------------------
		// INVALID JSON BODY
		// ------------------------------
		{
			name:           "invalid JSON body",
			method:         http.MethodPost,
			body:           `{ "envelope": "not-an-object" `,
			wantStatusCode: http.StatusBadRequest,
		},

		// ------------------------------
		// INVALID ENVELOPE STRUCTURE
		// ------------------------------
		{
			name:   "invalid envelope structure",
			method: http.MethodPost,
			body: `{
                "envelope": 123,
                "payload": {}
            }`,
			wantStatusCode: http.StatusUnprocessableEntity,
		},

		// ------------------------------
		// MISSING INDIVIDUAL FIELDS
		// ------------------------------
		{
			name:   "missing event_id",
			method: http.MethodPost,
			body: `{
                "envelope": {
                    "event_id": "",
                    "event_type": "claims.test.created",
                    "produced_at": "2026-05-01T14:23:00Z",
                    "source_system": "unit-test"
                },
                "payload": {}
            }`,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:   "missing event_type",
			method: http.MethodPost,
			body: `{
                "envelope": {
                    "event_id": "abc",
                    "event_type": "",
                    "produced_at": "2026-05-01T14:23:00Z",
                    "source_system": "unit-test"
                },
                "payload": {}
            }`,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:   "missing produced_at",
			method: http.MethodPost,
			body: `{
                "envelope": {
                    "event_id": "abc",
                    "event_type": "claims.test.created",
                    "produced_at": "",
                    "source_system": "unit-test"
                },
                "payload": {}
            }`,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:   "missing source_system",
			method: http.MethodPost,
			body: `{
                "envelope": {
                    "event_id": "abc",
                    "event_type": "claims.test.created",
                    "produced_at": "2026-05-01T14:23:00Z",
                    "source_system": ""
                },
                "payload": {}
            }`,
			wantStatusCode: http.StatusUnprocessableEntity,
		},

		// ------------------------------
		// MISSING ALL FIELDS
		// ------------------------------
		{
			name:   "missing all envelope fields",
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
