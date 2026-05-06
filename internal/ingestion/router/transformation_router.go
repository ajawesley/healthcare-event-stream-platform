package router

import (
	"context"
	"fmt"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/transformer"
	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
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

	return t, nil
}
