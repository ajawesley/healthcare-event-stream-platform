package router

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/compliance"
	"github.com/ajawes/hesp/internal/ingestion/detector"
	"github.com/ajawes/hesp/internal/ingestion/dispatcher"
	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/ingestion/pipeline"
	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

type FormatRouter struct {
	detector        detector.Detector
	normalizer      NormalizationRouter
	transformer     TransformationRouter
	complianceGuard compliance.ComplianceGuard
	dispatcher      dispatcher.Dispatcher
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

func WithComplianceGuard(g compliance.ComplianceGuard) Option {
	return func(r *FormatRouter) { r.complianceGuard = g }
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

func (r *FormatRouter) SetDetector(d detector.Detector) {
	r.detector = d
}

func (r *FormatRouter) SetNormalizer(n NormalizationRouter) {
	r.normalizer = n
}

func (r *FormatRouter) SetTransformer(t TransformationRouter) {
	r.transformer = t
}

func (r *FormatRouter) SetComplianceGuard(g compliance.ComplianceGuard) {
	r.complianceGuard = g
}

func (r *FormatRouter) SetDispatcher(d dispatcher.Dispatcher) {
	r.dispatcher = d
}

// -----------------------------------------------------------------------------
// Route(ctx, raw, env)
// -----------------------------------------------------------------------------
func (r *FormatRouter) Route(ctx context.Context, raw []byte, env api.Envelope) (*models.CanonicalEvent, error) {
	log := observability.WithTrace(ctx)

	// 0. PRE-SANITIZE
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

	// 1. DETECT FORMAT
	if r.detector == nil {
		log.Error("detector_undefined_error",
			zap.String("event_id", env.EventID),
		)
		observability.IncrementLineageFailure(ctx, "detect", "detector_not_configured")
		panic(errors.New("no detector configured"))
	}
	format := r.detector.Detect(sanitized)

	log.Debug("router_detect",
		zap.String("event_id", env.EventID),
		zap.String("format", string(format)),
	)

	// 2. LOOKUP NORMALIZER
	if r.normalizer == nil {
		log.Error("normalizer_undefined_error",
			zap.String("event_id", env.EventID),
		)
		observability.IncrementLineageFailure(ctx, "normalize", "normalizer_not_configured")
		panic(errors.New("no normalizer configured"))
	}

	normer, err := r.normalizer.NormalizerFor(format)
	if err != nil {
		log.Error("router_normalizer_lookup_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		observability.IncrementLineageFailure(ctx, "normalize", "normalizer_lookup_failed")
		return nil, fmt.Errorf("normalizer lookup failed: %w", err)
	}

	// 3. NORMALIZE (with metrics)
	normStart := time.Now()
	norm, err := normer.Normalize(ctx, sanitized, env)
	normDuration := time.Since(normStart)

	observability.ObserveStageLatency(ctx, "normalize", env.EventType, env.SourceSystem, normDuration)

	if err != nil {
		log.Error("router_normalize_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		observability.IncrementLineageFailure(ctx, "normalize", "normalization_failed")
		return nil, fmt.Errorf("normalization failed: %w", err)
	}

	log.Debug("router_normalize",
		zap.String("event_id", env.EventID),
		zap.Any("normalized", norm),
	)

	// 4. LOOKUP TRANSFORMER
	if r.transformer == nil {
		log.Error("transformer_undefined_error",
			zap.String("event_id", env.EventID),
		)
		observability.IncrementLineageFailure(ctx, "transform", "transformer_not_configured")
		panic(errors.New("no transformer configured"))
	}

	xform, err := r.transformer.TransformerFor(format)
	if err != nil {
		log.Error("router_transformer_lookup_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		observability.IncrementLineageFailure(ctx, "transform", "transformer_lookup_failed")
		return nil, fmt.Errorf("transformer lookup failed: %w", err)
	}

	// 5. TRANSFORM (with metrics)
	xformStart := time.Now()
	canon, err := xform.Transform(ctx, norm, env)
	xformDuration := time.Since(xformStart)

	observability.ObserveStageLatency(ctx, "transform", env.EventType, env.SourceSystem, xformDuration)

	if err != nil {
		log.Error("router_transform_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		observability.IncrementLineageFailure(ctx, "transform", "transformation_failed")
		return nil, fmt.Errorf("transformation failed: %w", err)
	}

	log.Debug("router_transform",
		zap.String("event_id", env.EventID),
		zap.Any("canonical", canon),
	)

	// -------------------------------------------------------------------------
	// 6. ⭐ COMPLIANCE GUARD (NEW FORMAL STAGE)
	// -------------------------------------------------------------------------
	if r.complianceGuard == nil {
		log.Error("compliance_guard_undefined_error",
			zap.String("event_id", env.EventID),
		)
		observability.IncrementLineageFailure(ctx, "compliance", "compliance_guard_not_configured")
		panic(errors.New("no compliance guard configured"))
	}

	if r.complianceGuard != nil {
		compStart := time.Now()
		err := r.complianceGuard.Apply(ctx, canon)
		compDuration := time.Since(compStart)

		observability.ObserveStageLatency(ctx, "compliance", env.EventType, env.SourceSystem, compDuration)

		if err != nil {
			log.Error("router_compliance_error",
				zap.String("event_id", env.EventID),
				zap.Error(err),
			)
			observability.IncrementLineageFailure(ctx, "compliance", "compliance_stage_failed")
			return nil, fmt.Errorf("compliance stage failed: %w", err)
		}

		// Lineage logging for compliance
		log.Debug("router_compliance",
			zap.String("event_id", env.EventID),
			zap.Bool("compliance_applied", canon.ComplianceApplied),
			zap.Bool("compliance_flag", canon.ComplianceFlag),
			zap.String("compliance_reason", canon.ComplianceReason),
			zap.String("compliance_rule_type", canon.ComplianceRuleType),
			zap.String("compliance_rule_id", canon.ComplianceRuleID),
		)
	}

	// 7. DISPATCH (with metrics)
	if r.dispatcher == nil {
		log.Error("dispatcher_undefined_error",
			zap.String("event_id", env.EventID),
		)
		observability.IncrementLineageFailure(ctx, "dispatch", "dispatcher_not_configured")
		panic(errors.New("no dispatcher configured"))
	}

	if r.dispatcher == nil {
		observability.IncrementLineageFailure(ctx, "dispatch", "dispatcher_not_configured")
		return nil, fmt.Errorf("dispatcher not configured")
	}

	dispatchStart := time.Now()
	err = r.dispatcher.Dispatch(ctx, canon, env, sanitized)
	dispatchDuration := time.Since(dispatchStart)

	observability.ObserveStageLatency(ctx, "dispatch", env.EventType, env.SourceSystem, dispatchDuration)

	if err != nil {
		log.Error("router_dispatch_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		observability.IncrementLineageFailure(ctx, "dispatch", "dispatch_failed")
		return nil, fmt.Errorf("dispatch failed: %w", err)
	}

	log.Debug("router_dispatch_success",
		zap.String("event_id", env.EventID),
	)

	// Successful end-to-end lineage event
	observability.IncrementLineageEvent(ctx, "dispatch", env.SourceSystem, env.EventType)

	return canon, nil
}
