package ingestion

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type ingestRequest struct {
	Envelope json.RawMessage `json:"envelope"`
	Payload  json.RawMessage `json:"payload"`
}

type envelope struct {
	EventID      string `json:"event_id"`
	EventType    string `json:"event_type"`
	ProducedAt   string `json:"produced_at"`
	SourceSystem string `json:"source_system"`
}

type ingestResponse struct {
	EventID    string `json:"event_id"`
	IngestedAt string `json:"ingested_at"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ingestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate envelope
	var env envelope
	if err := json.Unmarshal(req.Envelope, &env); err != nil {
		http.Error(w, "invalid envelope", http.StatusUnprocessableEntity)
		return
	}

	invalid := []string{}
	if env.EventID == "" {
		invalid = append(invalid, "envelope.event_id")
	}
	if env.EventType == "" {
		invalid = append(invalid, "envelope.event_type")
	}
	if env.ProducedAt == "" {
		invalid = append(invalid, "envelope.produced_at")
	}
	if env.SourceSystem == "" {
		invalid = append(invalid, "envelope.source_system")
	}

	if len(invalid) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":  "envelope_validation_failed",
			"fields": invalid,
		})
		return
	}

	// MVP: no S3 write yet — stub only
	ingestedAt := time.Now().UTC().Format(time.RFC3339)

	log.Printf(`{"event_id":"%s","event_type":"%s","source_system":"%s","ingested_at":"%s","outcome":"accepted","duration_ms":%d}`,
		env.EventID, env.EventType, env.SourceSystem, ingestedAt, time.Since(start).Milliseconds())

	resp := ingestResponse{
		EventID:    env.EventID,
		IngestedAt: ingestedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(resp)
}
