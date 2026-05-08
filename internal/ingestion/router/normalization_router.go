package router

import (
	"context"
	"fmt"
	"time"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/ingestion/normalizer"
	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

// -----------------------------------------------------------------------------
// Public NormalizationRouter interface
// -----------------------------------------------------------------------------

type NormalizationRouter interface {
	NormalizerFor(format config.Format) (normalizer.Normalizer, error)
}

// -----------------------------------------------------------------------------
// Internal implementation
// -----------------------------------------------------------------------------

type normalizationRouterImpl struct {
	normalizers map[config.Format]normalizer.Normalizer
}

func NewNormalizationRouter() NormalizationRouter {
	return &normalizationRouterImpl{
		normalizers: map[config.Format]normalizer.Normalizer{
			config.FormatHL7:     normalizer.NewHL7Normalizer(),
			config.FormatX12:     normalizer.NewX12Normalizer(),
			config.FormatFHIR:    normalizer.NewFHIRNormalizer(),
			config.FormatGeneric: normalizer.NewGenericNormalizer(),
		},
	}
}

// -----------------------------------------------------------------------------
// ⭐ lineageNormalizer decorator
// -----------------------------------------------------------------------------

type lineageNormalizer struct {
	inner normalizer.Normalizer
}

func (ln *lineageNormalizer) Normalize(ctx context.Context, raw []byte, env api.Envelope) (*models.NormalizedEvent, error) {
	stageStart := time.Now()

	norm, err := ln.inner.Normalize(ctx, raw, env)
	if err != nil {
		return nil, err
	}

	// -------------------------------------------------------------------------
	// ⭐ Mark normalization stage + log + metric
	// -------------------------------------------------------------------------
	if lineage := observability.GetLineage(ctx); lineage != nil {
		lineage.MarkStage("normalized")

		// Lineage log
		observability.Info(ctx, "lineage_stage_normalized",
			zap.String("event_id", lineage.EventID),
			zap.String("trace_id", lineage.TraceID),
			zap.Any("stages", lineage.Stages()),
		)

		// Lineage metric
		observability.ObserveLineageLatency(ctx, "normalized", stageStart)
	}

	return norm, nil
}

// -----------------------------------------------------------------------------
// ⭐ NormalizerFor now wraps normalizers with lineageNormalizer
// -----------------------------------------------------------------------------

func (r *normalizationRouterImpl) NormalizerFor(format config.Format) (normalizer.Normalizer, error) {
	ctx := context.Background()
	log := observability.WithTrace(ctx)

	n, ok := r.normalizers[format]
	if !ok {
		err := fmt.Errorf("no normalizer registered for format %s", format)

		log.Error("normalizer_lookup_error",
			zap.String("format", string(format)),
			zap.Error(err),
		)

		return nil, err
	}

	log.Debug("normalizer_lookup_success",
		zap.String("format", string(format)),
		zap.String("normalizer", fmt.Sprintf("%T", n)),
	)

	// ⭐ Wrap normalizer with lineage behavior
	return &lineageNormalizer{inner: n}, nil
}
