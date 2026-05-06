package router

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/detector"
	"github.com/ajawes/hesp/internal/ingestion/dispatcher"
	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/ingestion/pipeline"
	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
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

// -----------------------------------------------------------------------------
// ⭐ Correct signature: Route(raw, env)
// -----------------------------------------------------------------------------
func (r *FormatRouter) Route(raw []byte, env api.Envelope) (*models.CanonicalEvent, error) {
	ctx := context.Background()
	log := observability.WithTrace(ctx)

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

	log.Debug("router_presanitize",
		zap.String("event_id", env.EventID),
		zap.String("raw", rawQuoted),
		zap.String("sanitized", sanitizedQuoted),
	)

	// ----------------------------------------------------------------------
	// 1. DETECT FORMAT
	// ----------------------------------------------------------------------
	format := r.detector.Detect(sanitized)

	log.Debug("router_detect",
		zap.String("event_id", env.EventID),
		zap.String("format", string(format)),
	)

	// ----------------------------------------------------------------------
	// 2. LOOKUP NORMALIZER
	// ----------------------------------------------------------------------
	normer, err := r.normalizer.NormalizerFor(format)
	if err != nil {
		log.Error("router_normalizer_lookup_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("normalizer lookup failed: %w", err)
	}

	log.Debug("router_normalizer_lookup",
		zap.String("event_id", env.EventID),
		zap.String("normalizer", fmt.Sprintf("%T", normer)),
	)

	// ----------------------------------------------------------------------
	// 3. NORMALIZE
	// ----------------------------------------------------------------------
	norm, err := normer.Normalize(sanitized, env)
	if err != nil {
		log.Error("router_normalize_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("normalization failed: %w", err)
	}

	log.Debug("router_normalize",
		zap.String("event_id", env.EventID),
		zap.Any("normalized", norm),
	)

	// ----------------------------------------------------------------------
	// 4. LOOKUP TRANSFORMER
	// ----------------------------------------------------------------------
	xform, err := r.transformer.TransformerFor(format)
	if err != nil {
		log.Error("router_transformer_lookup_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("transformer lookup failed: %w", err)
	}

	log.Debug("router_transformer_lookup",
		zap.String("event_id", env.EventID),
		zap.String("transformer", fmt.Sprintf("%T", xform)),
	)

	// ----------------------------------------------------------------------
	// 5. TRANSFORM
	// ----------------------------------------------------------------------
	canon, err := xform.Transform(norm, env)
	if err != nil {
		log.Error("router_transform_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("transformation failed: %w", err)
	}

	log.Debug("router_transform",
		zap.String("event_id", env.EventID),
		zap.Any("canonical", canon),
	)

	// ----------------------------------------------------------------------
	// 6. DISPATCH
	// ----------------------------------------------------------------------
	if r.dispatcher == nil {
		return nil, fmt.Errorf("dispatcher not configured")
	}

	err = r.dispatcher.Dispatch(canon, env, sanitized)
	if err != nil {
		log.Error("router_dispatch_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("dispatch failed: %w", err)
	}

	log.Debug("router_dispatch_success",
		zap.String("event_id", env.EventID),
	)

	return canon, nil
}
