package normalizer

import (
	"context"
	"errors"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

type GenericNormalizer struct{}

func NewGenericNormalizer() *GenericNormalizer {
	return &GenericNormalizer{}
}

func (n *GenericNormalizer) Normalize(_ context.Context, raw []byte, env api.Envelope) (*models.NormalizedEvent, error) {
	ctx := context.Background()
	log := observability.WithTrace(ctx).With(
		zap.String("component", "generic_normalizer"),
		zap.String("event_id", env.EventID),
	)

	log.Info("generic_normalization_start")

	if raw == nil {
		err := errors.New("nil raw payload")
		log.Error("generic_normalization_error", zap.Error(err))
		return nil, err
	}

	ne := models.NewNormalizedEvent(config.FormatGeneric, raw)

	// Tests expect raw to be stored as a string
	ne.Fields["raw"] = string(raw)

	log.Info("generic_normalization_complete",
		zap.Any("fields", ne.Fields),
	)

	return ne, nil
}
