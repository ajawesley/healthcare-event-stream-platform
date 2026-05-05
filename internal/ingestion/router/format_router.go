package router

import (
	"fmt"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/detector"
	"github.com/ajawes/hesp/internal/ingestion/dispatcher"
	"github.com/ajawes/hesp/internal/ingestion/models"
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
	// 1. Detect format
	format := r.detector.Detect(raw)

	// 2. Lookup normalizer
	normer, err := r.normalizer.NormalizerFor(format)
	if err != nil {
		return nil, fmt.Errorf("normalizer lookup failed: %w", err)
	}

	// 3. Normalize (Normalize(raw, env))
	norm, err := normer.Normalize(raw, env)
	if err != nil {
		return nil, fmt.Errorf("normalization failed: %w", err)
	}

	// 4. Lookup transformer
	xform, err := r.transformer.TransformerFor(format)
	if err != nil {
		return nil, fmt.Errorf("transformer lookup failed: %w", err)
	}

	// 5. Transform (Transform(norm, env))
	canon, err := xform.Transform(norm, env)
	if err != nil {
		return nil, fmt.Errorf("transformation failed: %w", err)
	}

	// 6. Dispatch
	if r.dispatcher == nil {
		return nil, fmt.Errorf("dispatcher not configured")
	}

	if err := r.dispatcher.Dispatch(canon, env, raw); err != nil {
		return nil, fmt.Errorf("dispatch failed: %w", err)
	}

	return canon, nil
}
