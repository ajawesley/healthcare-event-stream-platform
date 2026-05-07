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

var ErrX12Unsupported = errors.New("unsupported X12 normalized event")

type X12Transformer struct{}

func NewX12Transformer() *X12Transformer {
	return &X12Transformer{}
}

func (t *X12Transformer) Transform(_ context.Context, ne *models.NormalizedEvent, env api.Envelope) (*models.CanonicalEvent, error) {
	ctx := context.Background()
	log := observability.WithTrace(ctx).With(
		zap.String("component", "transformer"),
		zap.String("transformer", "x12"),
		zap.String("event_id", env.EventID),
		zap.String("source_system", env.SourceSystem),
		zap.String("format", string(ne.Format)),
	)

	log.Info("x12_transform_start")

	if ne.Format != config.FormatX12 {
		err := fmt.Errorf("unsupported format: %s", ne.Format)
		log.Error("x12_transform_wrong_format", zap.Error(err))
		return nil, err
	}

	ce := &models.CanonicalEvent{
		EventID:      env.EventID,
		SourceSystem: env.SourceSystem,
		Format:       config.FormatX12,
		Metadata:     ne.Metadata,
		RawValue:     ne.Raw,
	}

	// -------------------------
	// Patient
	// -------------------------
	if id, ok := ne.Fields["nm1.il.patient_id"].(string); ok {
		ce.Patient = &models.CanonicalPatient{
			ID:        id,
			FirstName: asString(ne.Fields["nm1.il.first_name"]),
			LastName:  asString(ne.Fields["nm1.il.last_name"]),
		}

		log.Debug("x12_transform_patient",
			zap.String("id", id),
			zap.String("first_name", ce.Patient.FirstName),
			zap.String("last_name", ce.Patient.LastName),
		)
	}

	// -------------------------
	// Encounter
	// -------------------------
	if enc, ok := ne.Fields["clm.encounter_id"].(string); ok {
		ce.Encounter = &models.CanonicalEncounter{
			ID: enc,
		}

		log.Debug("x12_transform_encounter",
			zap.String("encounter_id", enc),
		)
	}

	// -------------------------
	// Observation / transaction code
	// -------------------------
	if code, ok := ne.Fields["st.transaction_code"].(string); ok {
		ce.Observation = &models.CanonicalObservation{
			Code: code,
		}

		log.Debug("x12_transform_observation",
			zap.String("transaction_code", code),
		)
	}

	log.Info("x12_transform_complete")

	return ce, nil
}
