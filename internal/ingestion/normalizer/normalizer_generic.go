package normalizer

import (
	"errors"
	"log/slog"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

type GenericNormalizer struct{}

func NewGenericNormalizer() *GenericNormalizer {
	return &GenericNormalizer{}
}

func (n *GenericNormalizer) Normalize(raw []byte, env api.Envelope) (*models.NormalizedEvent, error) {
	logger := slog.Default().With(
		"component", "generic_normalizer",
		"event_id", env.EventID,
	)

	logger.Info("starting Generic normalization")

	if raw == nil {
		return nil, errors.New("nil raw payload")
	}

	ne := models.NewNormalizedEvent(config.FormatGeneric, raw)

	// Tests expect raw to be stored as a string
	ne.Fields["raw"] = string(raw)

	logger.Info("Generic normalization complete")
	return ne, nil
}
