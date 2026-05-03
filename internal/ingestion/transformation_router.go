package ingestion

import "fmt"

// TransformationRouter maps Format → Transformer.
type TransformationRouter interface {
	TransformerFor(format Format) (Transformer, error)
}

type transformationRouterImpl struct {
	transformers map[Format]Transformer
}

func NewTransformationRouter() TransformationRouter {
	return &transformationRouterImpl{
		transformers: map[Format]Transformer{
			FormatHL7:     NewHL7Transformer(),
			FormatX12:     NewX12Transformer(),
			FormatFHIR:    NewFHIRTransformer(),
			FormatGeneric: NewGenericTransformer(),
		},
	}
}

func (r *transformationRouterImpl) TransformerFor(format Format) (Transformer, error) {
	t, ok := r.transformers[format]
	if !ok {
		return nil, fmt.Errorf("no transformer registered for format %s", format)
	}
	return t, nil
}
