package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/detector"
	"github.com/ajawes/hesp/internal/ingestion/normalizer"
	"github.com/ajawes/hesp/internal/ingestion/router"
	"github.com/ajawes/hesp/internal/ingestion/transformer"
)

type Handler struct {
	router              router.Router
	detector            detector.Detector
	cfg                 detector.DetectorConfig
	transformerRouter   transformer.TransformationRouter
	normalizationRouter normalizer.NormalizationRouter
}

type HandlerOption func(*Handler)

func WithConfig(cfg detector.DetectorConfig) HandlerOption {
	return func(h *Handler) {
		h.cfg = cfg
	}
}

func WithDetector(d detector.Detector) HandlerOption {
	return func(h *Handler) {
		h.detector = d
	}
}

func WithRouter(r router.Router) HandlerOption {
	return func(h *Handler) {
		h.router = r
	}
}

func WithNormalizationRouter(n normalizer.NormalizationRouter) HandlerOption {
	return func(h *Handler) {
		h.normalizationRouter = n
	}
}

func WithTransformationRouter(tr transformer.TransformationRouter) HandlerOption {
	return func(h *Handler) {
		h.transformerRouter = tr
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
		h.detector = detector.NewDetector(h.cfg)
	}

	// Build router if not injected
	if h.router == nil {
		h.router = router.NewParserRouter(h.detector)
	}

	// Build transformation router if not injected
	if h.transformerRouter == nil {
		h.transformerRouter = transformer.NewTransformationRouter()
	}

	// Build normalization router if not injected
	if h.normalizationRouter == nil {
		h.normalizationRouter = normalizer.NewNormalizationRouter()
	}

	return h
}

func loadDetectorConfigFromFile(path string) (detector.DetectorConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return detector.DetectorConfig{}, err
	}

	var cfg detector.DetectorConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return detector.DetectorConfig{}, err
	}

	return cfg, nil
}

func defaultDetectorConfig() detector.DetectorConfig {
	return detector.DetectorConfig{
		Rules: []detector.DetectionRule{
			{Name: "hl7_msh_prefix", Format: detector.FormatHL7, Prefix: "MSH|"},
			{Name: "x12_isa_prefix", Format: detector.FormatX12, Prefix: "ISA*"},
			{Name: "fhir_resource_type", Format: detector.FormatFHIR, ContainsKey: "resourceType"},
		},
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	var env api.Envelope
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

	// Transformation placeholder:
	// At this stage of the ingestion pipeline, we only invoke the transformer
	// to validate that the transformation layer is wired correctly. The actual
	// normalization logic for HL7, X12, FHIR, and Generic formats will be added
	// in the next ingestion slice. For now, this ensures:
	//   - the correct transformer is selected based on detected format
	//   - the handler exercises the transformation path end‑to‑end
	//   - DI and error handling behave as expected
	//
	// Once real transformers are implemented, this block will produce canonical
	// normalized output instead of discarding the result.
	transformer, err := h.transformerRouter.TransformerFor(routed.Format)
	if err != nil {
		http.Error(w, "no transformer for format", http.StatusInternalServerError)
		return
	}

	_, err = transformer.Transform(routed.Value)
	if err != nil {
		http.Error(w, "transformation failed", http.StatusUnprocessableEntity)
		return
	}

	// Normalization placeholder:
	// Converts the transformed payload into a canonical event.
	// Real normalization logic will be implemented in the next slices.
	normalizer, err := h.normalizationRouter.NormalizerFor(routed.Format)
	if err != nil {
		http.Error(w, "no normalizer for format", http.StatusInternalServerError)
		return
	}

	canonical, err := normalizer.Normalize(routed.Value, env)
	if err != nil {
		http.Error(w, "normalization failed", http.StatusUnprocessableEntity)
		return
	}

	// canonical is not yet returned in the HTTP response — that will change later.
	_ = canonical

	log.Printf(`{"event_id":"%s","event_type":"%s","source_system":"%s","ingested_at":"%s","outcome":"accepted","format":"%s","duration_ms":%d}`,
		env.EventID, env.EventType, env.SourceSystem, ingestedAt, routed.Format, time.Since(start).Milliseconds())

	resp := api.IngestResponse{
		EventID:    env.EventID,
		IngestedAt: ingestedAt,
		Format:     string(routed.Format),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(resp)
}
