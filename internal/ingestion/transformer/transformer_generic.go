package transformer

import (
	"context"
	"errors"
	"fmt"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

var ErrGenericUnsupported = errors.New("unsupported generic normalized event")

type GenericTransformer struct{}

func NewGenericTransformer() *GenericTransformer {
	return &GenericTransformer{}
}

func (t *GenericTransformer) Transform(ne *models.NormalizedEvent, env api.Envelope) (*models.CanonicalEvent, error) {
	ctx := context.Background()
	log := observability.WithTrace(ctx).With(
		zap.String("component", "transformer"),
		zap.String("transformer", "generic"),
		zap.String("event_id", env.EventID),
		zap.String("source_system", env.SourceSystem),
		zap.String("format", string(ne.Format)),
	)

	log.Info("generic_transform_start")

	// ----------------------------------------------------------------------
	// 1. Validate format
	// ----------------------------------------------------------------------
	if ne.Format != config.FormatGeneric {
		err := fmt.Errorf("unsupported format: %s", ne.Format)
		log.Error("generic_transform_unsupported_format", zap.Error(err))
		return nil, err
	}

	// ----------------------------------------------------------------------
	// 2. Map fields
	// ----------------------------------------------------------------------
	log.Debug("generic_transform_mapping_fields",
		zap.Int("field_count", len(ne.Fields)),
	)

	ce := &models.CanonicalEvent{
		EventID:      env.EventID,
		SourceSystem: env.SourceSystem,
		Format:       config.FormatGeneric,
		Metadata: map[string]any{
			"generic_fields": ne.Fields,
		},
	}

	// ----------------------------------------------------------------------
	// 3. Complete
	// ----------------------------------------------------------------------
	log.Info("generic_transform_complete",
		zap.Int("metadata_fields", len(ce.Metadata)),
	)

	return ce, nil
}
