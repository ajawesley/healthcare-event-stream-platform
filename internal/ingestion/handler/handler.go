package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/router"
)

type Handler struct {
	router router.Router
}

type HandlerOption func(*Handler)

func WithRouter(r router.Router) HandlerOption {
	return func(h *Handler) {
		h.router = r
	}
}

func NewHandler(opts ...HandlerOption) *Handler {
	h := &Handler{}

	for _, opt := range opts {
		opt(h)
	}

	if h.router == nil {
		panic("router must be provided to handler")
	}

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	log.Printf(`ingest_request_received method=%s path=%s remote=%s`, r.Method, r.URL.Path, r.RemoteAddr)

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// --- Read raw body ---
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf(`ingest_error stage=read_body error="%v"`, err)
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	// --- Log preview ---
	bodyPreview := string(rawBody)
	/*if len(bodyPreview) > 500 {
		bodyPreview = bodyPreview[:500] + "...(truncated)"
	}*/
	log.Printf(`ingest_body_preview raw_body="%s"`, strconv.Quote(bodyPreview))

	// --- Decode JSON ---
	var req api.IngestRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		log.Printf(`ingest_error stage=json_decode error="%v"`, err)
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// --- Decode envelope ---
	var env api.Envelope
	if err := json.Unmarshal(req.Envelope, &env); err != nil {
		log.Printf(`ingest_error stage=envelope_decode error="%v" envelope="%s"`, err, string(req.Envelope))
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_envelope_structure",
		})
		return
	}

	// --- Validate envelope ---
	invalid := []string{}
	if env.EventID == "" {
		invalid = append(invalid, "envelope.event_id")
	}
	if env.EventType == "" {
		invalid = append(invalid, "envelope.event_type")
	}
	if env.ProducedAt.IsZero() {
		invalid = append(invalid, "envelope.produced_at (empty)")
	}
	if env.SourceSystem == "" {
		invalid = append(invalid, "envelope.source_system")
	}

	if len(invalid) > 0 {
		log.Printf(`ingest_error stage=envelope_validation fields=%v`, invalid)
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":  "envelope_validation_failed",
			"fields": invalid,
		})
		return
	}

	// --- Validate payload ---
	if len(req.Payload) == 0 {
		log.Printf(`ingest_error stage=payload_missing`)
		http.Error(w, "missing payload", http.StatusBadRequest)
		return
	}

	// --- Route ---
	canonical, err := h.router.Route(req.Payload, env)
	if err != nil {
		errMsg := fmt.Sprintf(`ingest_rejected event_id="%s" event_type="%s" error="%v"`, env.EventID, env.EventType, err)
		log.Printf(errMsg)
		http.Error(w, "unsupported or invalid payload: "+errMsg, http.StatusUnprocessableEntity)
		return
	}

	// --- Success ---
	log.Printf(`ingest_accepted event_id="%s" format="%s" duration_ms=%d`,
		env.EventID, canonical.Format, time.Since(start).Milliseconds())

	resp := api.IngestResponse{
		EventID:    env.EventID,
		IngestedAt: time.Now().UTC().Format(time.RFC3339),
		Format:     string(canonical.Format),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(resp)
}
