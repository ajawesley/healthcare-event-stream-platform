package transformer

import (
	"context"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

type TransformationRouter interface {
	TransformerFor(format config.Format) (Transformer, error)
}

type Transformer interface {
	Transform(ctx context.Context, norm *models.NormalizedEvent, env api.Envelope) (*models.CanonicalEvent, error)
}
