package normalizer

import (
	"fmt"

	"github.com/ajawes/hesp/internal/ingestion/detector"
)

// NormalizationRouter maps Format → Normalizer.
type NormalizationRouter interface {
	NormalizerFor(format detector.Format) (Normalizer, error)
}

type normalizationRouterImpl struct {
	normalizers map[detector.Format]Normalizer
}

func NewNormalizationRouter() NormalizationRouter {
	return &normalizationRouterImpl{
		normalizers: map[detector.Format]Normalizer{
			detector.FormatHL7:     NewHL7Normalizer(),
			detector.FormatX12:     NewX12Normalizer(),
			detector.FormatFHIR:    NewFHIRNormalizer(),
			detector.FormatGeneric: NewGenericNormalizer(),
		},
	}
}

func (r *normalizationRouterImpl) NormalizerFor(format detector.Format) (Normalizer, error) {
	n, ok := r.normalizers[format]
	if !ok {
		return nil, fmt.Errorf("no normalizer registered for format %s", format)
	}
	return n, nil
}
