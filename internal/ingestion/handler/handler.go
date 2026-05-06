package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/router"
	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
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
	ctx := r.Context()
	start := time.Now()

	log := observability.WithTrace(ctx)
	log.Info("ingest_request_received",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
	)

	if r.Method != http.MethodPost {
		observability.Error(ctx, "method_not_allowed", fmt.Errorf("invalid method"), "invalid_method", "only POST allowed",
			zap.String("method", r.Method),
		)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// --- Read raw body ---
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		observability.Error(ctx, "read_body_failed", err, "read_error", "failed to read request body")
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	// --- Log preview ---
	bodyPreview := string(rawBody)
	log.Info("ingest_body_preview",
		zap.String("raw_body", strconv.Quote(bodyPreview)),
	)

	// --- Decode JSON ---
	var req api.IngestRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		observability.Error(ctx, "json_decode_failed", err, "json_error", "failed to decode ingest request")
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// --- Decode envelope ---
	var env api.Envelope
	if err := json.Unmarshal(req.Envelope, &env); err != nil {
		observability.Error(ctx, "envelope_decode_failed", err, "envelope_error", "failed to decode envelope",
			zap.String("envelope_raw", string(req.Envelope)),
		)

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
		observability.Error(ctx, "envelope_validation_failed", fmt.Errorf("invalid envelope"), "validation_error", "envelope validation failed",
			zap.Strings("invalid_fields", invalid),
		)

		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":  "envelope_validation_failed",
			"fields": invalid,
		})
		return
	}

	// --- Validate payload ---
	if len(req.Payload) == 0 {
		observability.Error(ctx, "payload_missing", fmt.Errorf("missing payload"), "payload_error", "payload is empty")
		http.Error(w, "missing payload", http.StatusBadRequest)
		return
	}

	// --- Route ---
	canonical, err := h.router.Route(req.Payload, env)
	if err != nil {
		observability.Error(ctx, "ingest_rejected", err, "routing_error", "router rejected payload",
			zap.String("event_id", env.EventID),
			zap.String("event_type", env.EventType),
		)

		http.Error(w, "unsupported or invalid payload", http.StatusUnprocessableEntity)
		return
	}

	// --- Success ---
	log.Info("ingest_accepted",
		zap.String("event_id", env.EventID),
		zap.String("format", string(canonical.Format)),
		zap.Int64("duration_ms", time.Since(start).Milliseconds()),
	)

	resp := api.IngestResponse{
		EventID:    env.EventID,
		IngestedAt: time.Now().UTC().Format(time.RFC3339),
		Format:     string(canonical.Format),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(resp)
}
