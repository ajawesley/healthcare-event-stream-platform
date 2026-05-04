package normalizer

import (
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/detector"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

type genericNormalizer struct{}

func NewGenericNormalizer() Normalizer {
	return &genericNormalizer{}
}

func (n *genericNormalizer) Normalize(value any, meta api.Envelope) (*models.CanonicalEvent, error) {
	return &models.CanonicalEvent{
		EventID:      meta.EventID,
		SourceSystem: meta.SourceSystem,
		Format:       detector.FormatGeneric,
		RawValue:     value,
	}, nil
}
