package ingestion

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type Handler struct {
	router   Router
	detector Detector
	cfg      DetectorConfig
}

type HandlerOption func(*Handler)

func WithConfig(cfg DetectorConfig) HandlerOption {
	return func(h *Handler) {
		h.cfg = cfg
	}
}

func WithDetector(d Detector) HandlerOption {
	return func(h *Handler) {
		h.detector = d
	}
}

func WithRouter(r Router) HandlerOption {
	return func(h *Handler) {
		h.router = r
	}
}

func NewHandler(opts ...HandlerOption) *Handler {
	h := &Handler{}

	for _, opt := range opts {
		opt(h)
	}

	// Load config if not injected
	if h.cfg.Rules == nil {
		cfgPath := os.Getenv("INGESTION_DETECTION_CONFIG")

		if cfgPath != "" {
			loaded, err := loadDetectorConfigFromFile(cfgPath)
			if err != nil {
				log.Printf("failed to load detector config from %s, falling back to defaults: %v", cfgPath, err)
				h.cfg = defaultDetectorConfig()
			} else {
				h.cfg = loaded
			}
		} else {
			h.cfg = defaultDetectorConfig()
		}
	}

	// Build detector if not injected
	if h.detector == nil {
		h.detector = NewDetector(h.cfg)
	}

	// Build router if not injected
	if h.router == nil {
		h.router = NewRouter(h.detector)
	}

	return h
}

func loadDetectorConfigFromFile(path string) (DetectorConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return DetectorConfig{}, err
	}

	var cfg DetectorConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return DetectorConfig{}, err
	}

	return cfg, nil
}

func defaultDetectorConfig() DetectorConfig {
	return DetectorConfig{
		Rules: []DetectionRule{
			{Name: "hl7_msh_prefix", Format: FormatHL7, Prefix: "MSH|"},
			{Name: "x12_isa_prefix", Format: FormatX12, Prefix: "ISA*"},
			{Name: "fhir_resource_type", Format: FormatFHIR, ContainsKey: "resourceType"},
		},
	}
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
	Format     string `json:"format"`
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

	var env envelope
	if err := json.Unmarshal(req.Envelope, &env); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_envelope_structure",
		})
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
		invalid = append(invalid, "envelope.produced_at (empty)")
	} else if _, err := time.Parse(time.RFC3339, env.ProducedAt); err != nil {
		invalid = append(invalid, "envelope.produced_at (invalid format, must be RFC3339)")
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

	if len(req.Payload) == 0 {
		http.Error(w, "missing payload", http.StatusBadRequest)
		return
	}

	routed, err := h.router.Route(req.Payload)
	if err != nil {
		log.Printf(`{"event_id":"%s","event_type":"%s","source_system":"%s","outcome":"rejected","error":"%s"}`,
			env.EventID, env.EventType, env.SourceSystem, err.Error())
		http.Error(w, "unsupported or invalid payload", http.StatusUnprocessableEntity)
		return
	}

	ingestedAt := time.Now().UTC().Format(time.RFC3339)

	log.Printf(`{"event_id":"%s","event_type":"%s","source_system":"%s","ingested_at":"%s","outcome":"accepted","format":"%s","duration_ms":%d}`,
		env.EventID, env.EventType, env.SourceSystem, ingestedAt, routed.Format, time.Since(start).Milliseconds())

	resp := ingestResponse{
		EventID:    env.EventID,
		IngestedAt: ingestedAt,
		Format:     string(routed.Format),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(resp)
}
