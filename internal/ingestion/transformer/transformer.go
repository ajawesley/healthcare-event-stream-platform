package transformer

import (
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

type Transformer interface {
	Transform(ne *models.NormalizedEvent, env api.Envelope) (*models.CanonicalEvent, error)
}
