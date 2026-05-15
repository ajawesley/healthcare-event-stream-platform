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
	"go.opentelemetry.io/otel"
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

// -----------------------------------------------------------------------------
// Setter Methods (KEPT as requested)
// -----------------------------------------------------------------------------
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
	tr := otel.Tracer("router")
	ctx, span := tr.Start(ctx, "format_router")
	defer span.End()

	log := observability.WithTrace(ctx)

	// 0. PRE-SANITIZE
	_, preSpan := tr.Start(ctx, "router_presanitize")
	sanitized := pipeline.PreSanitize(raw)

	rawQuoted := strconv.Quote(string(raw))
	sanitizedQuoted := strconv.Quote(string(sanitized))

	if len(rawQuoted) > 4000 {
		rawQuoted = rawQuoted[:4000] + `"...(truncated)`
	}
	if len(sanitizedQuoted) > 4000 {
		sanitizedQuoted = sanitizedQuoted[:4000] + `"...(truncated)`
	}

	log.Info("router_presanitize",
		zap.String("event_id", env.EventID),
		zap.String("raw", rawQuoted),
		zap.String("sanitized", sanitizedQuoted),
	)
	preSpan.End()

	// 1. DETECT FORMAT
	detectCtx, detectSpan := tr.Start(ctx, "router_detect")
	if r.detector == nil {
		log.Error("detector_undefined_error",
			zap.String("event_id", env.EventID),
		)
		observability.IncrementLineageFailure(detectCtx, "detect", "detector_not_configured")
		detectSpan.End()
		panic(errors.New("no detector configured"))
	}
	format := r.detector.Detect(detectCtx, sanitized)

	log.Info("router_detect",
		zap.String("event_id", env.EventID),
		zap.String("format", string(format)),
	)
	detectSpan.End()

	// 2. LOOKUP NORMALIZER
	normLookupCtx, normLookupSpan := tr.Start(ctx, "router_normalizer_lookup")
	if r.normalizer == nil {
		log.Error("normalizer_undefined_error",
			zap.String("event_id", env.EventID),
		)
		observability.IncrementLineageFailure(normLookupCtx, "normalize", "normalizer_not_configured")
		normLookupSpan.End()
		panic(errors.New("no normalizer configured"))
	}

	normer, err := r.normalizer.NormalizerFor(format)
	if err != nil {
		log.Error("router_normalizer_lookup_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		observability.IncrementLineageFailure(normLookupCtx, "normalize", "normalizer_lookup_failed")
		normLookupSpan.End()
		return nil, fmt.Errorf("normalizer lookup failed: %w", err)
	}
	normLookupSpan.End()

	// 3. NORMALIZE
	normCtx, normSpan := tr.Start(ctx, "router_normalize")
	normStart := time.Now()
	norm, err := normer.Normalize(normCtx, sanitized, env)
	normDuration := time.Since(normStart)

	observability.ObserveStageLatency(normCtx, "normalize", env.EventType, env.SourceSystem, normDuration)

	if err != nil {
		log.Error("router_normalize_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		observability.IncrementLineageFailure(normCtx, "normalize", "normalization_failed")
		normSpan.End()
		return nil, fmt.Errorf("normalization failed: %w", err)
	}

	log.Info("router_normalize",
		zap.String("event_id", env.EventID),
		zap.Any("normalized", norm),
	)
	normSpan.End()

	// 4. LOOKUP TRANSFORMER
	xformLookupCtx, xformLookupSpan := tr.Start(ctx, "router_transformer_lookup")
	if r.transformer == nil {
		log.Error("transformer_undefined_error",
			zap.String("event_id", env.EventID),
		)
		observability.IncrementLineageFailure(xformLookupCtx, "transform", "transformer_not_configured")
		xformLookupSpan.End()
		panic(errors.New("no transformer configured"))
	}

	xform, err := r.transformer.TransformerFor(xformLookupCtx, format)
	if err != nil {
		log.Error("router_transformer_lookup_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		observability.IncrementLineageFailure(xformLookupCtx, "transform", "transformer_lookup_failed")
		xformLookupSpan.End()
		return nil, fmt.Errorf("transformer lookup failed: %w", err)
	}
	xformLookupSpan.End()

	// 5. TRANSFORM
	xformCtx, xformSpan := tr.Start(ctx, "router_transform")
	xformStart := time.Now()
	canon, err := xform.Transform(xformCtx, norm, env)
	xformDuration := time.Since(xformStart)

	observability.ObserveStageLatency(xformCtx, "transform", env.EventType, env.SourceSystem, xformDuration)

	if err != nil {
		log.Error("router_transform_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		observability.IncrementLineageFailure(xformCtx, "transform", "transformation_failed")
		xformSpan.End()
		return nil, fmt.Errorf("transformation failed: %w", err)
	}

	log.Info("router_transform",
		zap.String("event_id", env.EventID),
		zap.Any("canonical", canon),
	)
	xformSpan.End()

	// 6. COMPLIANCE GUARD
	compCtx, compSpan := tr.Start(ctx, "router_compliance")
	if r.complianceGuard == nil {
		log.Error("compliance_guard_undefined_error",
			zap.String("event_id", env.EventID),
		)
		observability.IncrementLineageFailure(compCtx, "compliance", "compliance_guard_not_configured")
		compSpan.End()
		panic(errors.New("no compliance guard configured"))
	}

	compStart := time.Now()
	err = r.complianceGuard.Apply(compCtx, canon)
	compDuration := time.Since(compStart)

	observability.ObserveStageLatency(compCtx, "compliance", env.EventType, env.SourceSystem, compDuration)

	if err != nil {
		log.Error("router_compliance_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		observability.IncrementLineageFailure(compCtx, "compliance", "compliance_stage_failed")
		compSpan.End()
		return nil, fmt.Errorf("compliance stage failed: %w", err)
	}

	log.Debug("router_compliance",
		zap.String("event_id", env.EventID),
		zap.Bool("compliance_applied", canon.ComplianceApplied),
		zap.Bool("compliance_flag", canon.ComplianceFlag),
		zap.String("compliance_reason", canon.ComplianceReason),
		zap.String("compliance_rule_type", canon.ComplianceRuleType),
		zap.String("compliance_rule_id", canon.ComplianceRuleID),
	)
	compSpan.End()

	// 7. DISPATCH
	dispatchCtx, dispatchSpan := tr.Start(ctx, "router_dispatch")
	if r.dispatcher == nil {
		log.Error("dispatcher_undefined_error",
			zap.String("event_id", env.EventID),
		)
		observability.IncrementLineageFailure(dispatchCtx, "dispatch", "dispatcher_not_configured")
		dispatchSpan.End()
		panic(errors.New("no dispatcher configured"))
	}

	dispatchStart := time.Now()
	err = r.dispatcher.Dispatch(dispatchCtx, canon, env, sanitized)
	dispatchDuration := time.Since(dispatchStart)

	observability.ObserveStageLatency(dispatchCtx, "dispatch", env.EventType, env.SourceSystem, dispatchDuration)

	if err != nil {
		log.Error("router_dispatch_error",
			zap.String("event_id", env.EventID),
			zap.Error(err),
		)
		observability.IncrementLineageFailure(dispatchCtx, "dispatch", "dispatch_failed")
		dispatchSpan.End()
		return nil, fmt.Errorf("dispatch failed: %w", err)
	}

	log.Debug("router_dispatch_success",
		zap.String("event_id", env.EventID),
	)
	dispatchSpan.End()

	observability.IncrementLineageEvent(ctx, "dispatch", env.SourceSystem, env.EventType)

	return canon, nil
}
