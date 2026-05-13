package router

import (
	"context"
	"fmt"
	"time"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/ingestion/transformer"
	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

// -----------------------------------------------------------------------------
// Public TransformationRouter interface
// -----------------------------------------------------------------------------

type TransformationRouter interface {
	TransformerFor(format config.Format) (transformer.Transformer, error)
}

// -----------------------------------------------------------------------------
// Internal implementation
// -----------------------------------------------------------------------------

type transformationRouterImpl struct {
	transformers map[config.Format]transformer.Transformer
}

func NewTransformationRouter() TransformationRouter {
	return &transformationRouterImpl{
		transformers: map[config.Format]transformer.Transformer{
			config.FormatHL7:     transformer.NewHL7Transformer(),
			config.FormatX12:     transformer.NewX12Transformer(),
			config.FormatFHIR:    transformer.NewFHIRTransformer(),
			config.FormatGeneric: transformer.NewGenericTransformer(),
		},
	}
}

// -----------------------------------------------------------------------------
// ⭐ lineageTransformer decorator
// -----------------------------------------------------------------------------

type lineageTransformer struct {
	inner transformer.Transformer
}

func (lt *lineageTransformer) Transform(ctx context.Context, norm *models.NormalizedEvent, env api.Envelope) (*models.CanonicalEvent, error) {
	stageStart := time.Now()

	canon, err := lt.inner.Transform(ctx, norm, env)
	if err != nil {
		return nil, err
	}

	// -------------------------------------------------------------------------
	// ⭐ Mark canonicalization stage + log + metric
	// -------------------------------------------------------------------------
	if lineage := observability.GetLineage(ctx); lineage != nil {
		lineage.MarkStage("canonicalized")

		// Lineage log
		observability.Info(ctx, "lineage_stage_canonicalized",
			zap.String("event_id", lineage.EventID),
			zap.String("trace_id", lineage.TraceID),
			zap.Any("stages", lineage.Stages()),
		)

		// Lineage metric
		observability.ObserveLineageLatency(ctx, "canonicalized", stageStart)
	}

	return canon, nil
}

// -----------------------------------------------------------------------------
// ⭐ TransformerFor now wraps transformers with lineageTransformer
// -----------------------------------------------------------------------------

func (r *transformationRouterImpl) TransformerFor(format config.Format) (transformer.Transformer, error) {
	ctx := context.Background()
	log := observability.WithTrace(ctx)

	t, ok := r.transformers[format]
	if !ok {
		err := fmt.Errorf("no transformer registered for format %s", format)

		log.Error("transformer_lookup_error",
			zap.String("format", string(format)),
			zap.Error(err),
		)

		return nil, err
	}

	log.Debug("transformer_lookup_success",
		zap.String("format", string(format)),
		zap.String("transformer", fmt.Sprintf("%T", t)),
	)

	// ⭐ Wrap transformer with lineage behavior
	return &lineageTransformer{inner: t}, nil
}
