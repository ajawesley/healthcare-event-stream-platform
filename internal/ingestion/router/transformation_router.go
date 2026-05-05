package router

import (
	"fmt"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/transformer"
)

// TransformationRouter maps Format → Transformer.
type TransformationRouter interface {
	TransformerFor(format config.Format) (transformer.Transformer, error)
}

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

func (r *transformationRouterImpl) TransformerFor(format config.Format) (transformer.Transformer, error) {
	t, ok := r.transformers[format]
	if !ok {
		return nil, fmt.Errorf("no transformer registered for format %s", format)
	}
	return t, nil
}
