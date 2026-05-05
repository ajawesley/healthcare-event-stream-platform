package router

import (
	"fmt"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/normalizer"
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
	n, ok := r.normalizers[format]
	if !ok {
		return nil, fmt.Errorf("no normalizer registered for format %s", format)
	}
	return n, nil
}
