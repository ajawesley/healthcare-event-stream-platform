package transformer

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

var ErrGenericUnsupported = errors.New("unsupported generic normalized event")

type GenericTransformer struct{ logger *slog.Logger }

func NewGenericTransformer() *GenericTransformer {
	return &GenericTransformer{logger: slog.Default()}
}

func (t *GenericTransformer) Transform(ne *models.NormalizedEvent, env api.Envelope) (*models.CanonicalEvent, error) {
	log := t.logger.With(
		"component", "transformer",
		"transformer", "generic",
		"event_id", env.EventID,
		"source_system", env.SourceSystem,
		"format", ne.Format,
	)

	log.Info("transformer: start")

	if ne.Format != config.FormatGeneric {
		log.Error("transformer: unsupported format")
		return nil, fmt.Errorf("unsupported format: %s", ne.Format)
	}

	log.Debug("transformer: mapping generic fields", "field_count", len(ne.Fields))

	ce := &models.CanonicalEvent{
		EventID:      env.EventID,
		SourceSystem: env.SourceSystem,
		Format:       config.FormatGeneric,
		Metadata: map[string]any{
			"generic_fields": ne.Fields,
		},
	}

	log.Info("transformer: complete",
		"canonical_fields", len(ce.Metadata),
	)

	return ce, nil
}
