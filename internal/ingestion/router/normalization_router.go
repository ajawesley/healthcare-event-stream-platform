package router

import (
	"context"
	"fmt"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/normalizer"
	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

// NormalizationRouter maps Format → Normalizer.
type NormalizationRouter interface {
	NormalizerFor(format config.Format) (normalizer.Normalizer, error)
}

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

	return n, nil
}
