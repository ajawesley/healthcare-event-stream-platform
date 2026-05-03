package ingestion

import "fmt"

// NormalizationRouter maps Format → Normalizer.
type NormalizationRouter interface {
	NormalizerFor(format Format) (Normalizer, error)
}

type normalizationRouterImpl struct {
	normalizers map[Format]Normalizer
}

func NewNormalizationRouter() NormalizationRouter {
	return &normalizationRouterImpl{
		normalizers: map[Format]Normalizer{
			FormatHL7:     NewHL7Normalizer(),
			FormatX12:     NewX12Normalizer(),
			FormatFHIR:    NewFHIRNormalizer(),
			FormatGeneric: NewGenericNormalizer(),
		},
	}
}

func (r *normalizationRouterImpl) NormalizerFor(format Format) (Normalizer, error) {
	n, ok := r.normalizers[format]
	if !ok {
		return nil, fmt.Errorf("no normalizer registered for format %s", format)
	}
	return n, nil
}
