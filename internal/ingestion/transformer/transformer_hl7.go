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

var ErrHL7Unsupported = errors.New("unsupported HL7 normalized event")

type HL7Transformer struct{}

func NewHL7Transformer() *HL7Transformer {
	return &HL7Transformer{}
}

func (t *HL7Transformer) Transform(ne *models.NormalizedEvent, env api.Envelope) (*models.CanonicalEvent, error) {
	ctx := context.Background()
	log := observability.WithTrace(ctx).With(
		zap.String("component", "transformer"),
		zap.String("transformer", "hl7"),
		zap.String("event_id", env.EventID),
		zap.String("source_system", env.SourceSystem),
		zap.String("format", string(ne.Format)),
	)

	log.Info("hl7_transform_start")

	if ne.Format != config.FormatHL7 {
		err := fmt.Errorf("unsupported format: %s", ne.Format)
		log.Error("hl7_transform_wrong_format", zap.Error(err))
		return nil, err
	}

	ce := &models.CanonicalEvent{
		EventID:      env.EventID,
		SourceSystem: env.SourceSystem,
		Format:       config.FormatHL7,
		Metadata:     ne.Metadata,
		RawValue:     ne.Raw,
	}

	// -------------------------
	// Patient
	// -------------------------
	if id, ok := ne.Fields["pid.id"].(string); ok {
		ce.Patient = &models.CanonicalPatient{
			ID:        id,
			FirstName: asString(ne.Fields["pid.first_name"]),
			LastName:  asString(ne.Fields["pid.last_name"]),
		}

		log.Debug("hl7_transform_patient",
			zap.String("id", id),
			zap.String("first_name", ce.Patient.FirstName),
			zap.String("last_name", ce.Patient.LastName),
		)
	}

	// -------------------------
	// Encounter
	// -------------------------
	if et, ok := ne.Fields["pv1.encounter_type"].(string); ok {
		ce.Encounter = &models.CanonicalEncounter{
			Type: et,
		}

		log.Debug("hl7_transform_encounter",
			zap.String("encounter_type", et),
		)
	}

	// -------------------------
	// Observation
	// -------------------------
	if msgType, ok := ne.Fields["msh.message_type"].(string); ok {
		ce.Observation = &models.CanonicalObservation{
			Code: msgType,
		}

		log.Debug("hl7_transform_observation",
			zap.String("message_type", msgType),
		)
	}

	log.Info("hl7_transform_complete")

	return ce, nil
}
