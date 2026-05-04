package transformer

import (
	"fmt"

	"github.com/ajawes/hesp/internal/ingestion/detector"
)

// TransformationRouter maps Format → Transformer.
type TransformationRouter interface {
	TransformerFor(format detector.Format) (Transformer, error)
}

type transformationRouterImpl struct {
	transformers map[detector.Format]Transformer
}

func NewTransformationRouter() TransformationRouter {
	return &transformationRouterImpl{
		transformers: map[detector.Format]Transformer{
			detector.FormatHL7:     NewHL7Transformer(),
			detector.FormatX12:     NewX12Transformer(),
			detector.FormatFHIR:    NewFHIRTransformer(),
			detector.FormatGeneric: NewGenericTransformer(),
		},
	}
}

func (r *transformationRouterImpl) TransformerFor(format detector.Format) (Transformer, error) {
	t, ok := r.transformers[format]
	if !ok {
		return nil, fmt.Errorf("no transformer registered for format %s", format)
	}
	return t, nil
}
