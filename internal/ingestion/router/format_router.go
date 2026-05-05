package router

import (
	"fmt"
	"log"
	"strconv"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/detector"
	"github.com/ajawes/hesp/internal/ingestion/dispatcher"
	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/ingestion/pipeline"
)

type FormatRouter struct {
	detector    detector.Detector
	normalizer  NormalizationRouter
	transformer TransformationRouter
	dispatcher  dispatcher.Dispatcher
}

type Option func(*FormatRouter)

func WithDetector(d detector.Detector) Option {
	return func(r *FormatRouter) { r.detector = d }
}

func WithNormalizationRouter(n NormalizationRouter) Option {
	return func(r *FormatRouter) { r.normalizer = n }
}

func WithTransformationRouter(t TransformationRouter) Option {
	return func(r *FormatRouter) { r.transformer = t }
}

func WithDispatcher(d dispatcher.Dispatcher) Option {
	return func(r *FormatRouter) { r.dispatcher = d }
}

func NewFormatRouter(opts ...Option) *FormatRouter {
	r := &FormatRouter{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *FormatRouter) SetDispatcher(d dispatcher.Dispatcher) {
	r.dispatcher = d
}

// ⭐ Correct signature: Route(raw, env)
func (r *FormatRouter) Route(raw []byte, env api.Envelope) (*models.CanonicalEvent, error) {

	// ----------------------------------------------------------------------
	// 0. PRE-SANITIZE FIRST
	// ----------------------------------------------------------------------
	sanitized := pipeline.PreSanitize(raw)

	rawQuoted := strconv.Quote(string(raw))
	sanitizedQuoted := strconv.Quote(string(sanitized))

	if len(rawQuoted) > 4000 {
		rawQuoted = rawQuoted[:4000] + `"...(truncated)`
	}
	if len(sanitizedQuoted) > 4000 {
		sanitizedQuoted = sanitizedQuoted[:4000] + `"...(truncated)`
	}

	log.Printf(
		`router_stage="presanitize" event_id="%s" raw=%s sanitized=%s`,
		env.EventID,
		rawQuoted,
		sanitizedQuoted,
	)

	// ----------------------------------------------------------------------
	// 1. DETECT FORMAT (using sanitized payload)
	// ----------------------------------------------------------------------
	format := r.detector.Detect(sanitized)

	log.Printf(
		`router_stage="detect" event_id="%s" detected_format="%s" detector_input=%s`,
		env.EventID,
		format,
		sanitizedQuoted,
	)

	// ----------------------------------------------------------------------
	// 2. LOOKUP NORMALIZER
	// ----------------------------------------------------------------------
	normer, err := r.normalizer.NormalizerFor(format)
	if err != nil {
		log.Printf(`router_stage="normalizer_lookup_error" event_id="%s" error="%v"`, env.EventID, err)
		return nil, fmt.Errorf("normalizer lookup failed: %w", err)
	}

	log.Printf(
		`router_stage="normalizer_lookup" event_id="%s" normalizer="%T"`,
		env.EventID,
		normer,
	)

	// ----------------------------------------------------------------------
	// 3. NORMALIZE
	// ----------------------------------------------------------------------
	norm, err := normer.Normalize(sanitized, env)
	if err != nil {
		log.Printf(`router_stage="normalize_error" event_id="%s" error="%v"`, env.EventID, err)
		return nil, fmt.Errorf("normalization failed: %w", err)
	}

	log.Printf(
		`router_stage="normalize" event_id="%s" normalized="%+v"`,
		env.EventID,
		norm,
	)

	// ----------------------------------------------------------------------
	// 4. LOOKUP TRANSFORMER
	// ----------------------------------------------------------------------
	xform, err := r.transformer.TransformerFor(format)
	if err != nil {
		log.Printf(`router_stage="transformer_lookup_error" event_id="%s" error="%v"`, env.EventID, err)
		return nil, fmt.Errorf("transformer lookup failed: %w", err)
	}

	log.Printf(
		`router_stage="transformer_lookup" event_id="%s" transformer="%T"`,
		env.EventID,
		xform,
	)

	// ----------------------------------------------------------------------
	// 5. TRANSFORM
	// ----------------------------------------------------------------------
	canon, err := xform.Transform(norm, env)
	if err != nil {
		log.Printf(`router_stage="transform_error" event_id="%s" error="%v"`, env.EventID, err)
		return nil, fmt.Errorf("transformation failed: %w", err)
	}

	log.Printf(
		`router_stage="transform" event_id="%s" canonical="%+v"`,
		env.EventID,
		canon,
	)

	// ----------------------------------------------------------------------
	// 6. DISPATCH (using sanitized payload)
	// ----------------------------------------------------------------------
	if r.dispatcher == nil {
		return nil, fmt.Errorf("dispatcher not configured")
	}

	err = r.dispatcher.Dispatch(canon, env, sanitized)
	if err != nil {
		log.Printf(`router_stage="dispatch_error" event_id="%s" error="%v"`, env.EventID, err)
		return nil, fmt.Errorf("dispatch failed: %w", err)
	}

	log.Printf(
		`router_stage="dispatch" event_id="%s" status="success"`,
		env.EventID,
	)

	return canon, nil
}
