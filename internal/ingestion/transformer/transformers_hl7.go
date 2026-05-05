package transformer

import (
	"errors"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

var ErrHL7Unsupported = errors.New("unsupported HL7 normalized event")

type HL7Transformer struct{}

func NewHL7Transformer() *HL7Transformer {
	return &HL7Transformer{}
}

func (t *HL7Transformer) Transform(ne *models.NormalizedEvent, env api.Envelope) (*models.CanonicalEvent, error) {
	if ne.Format != config.FormatHL7 {
		return nil, ErrHL7Unsupported
	}

	ce := &models.CanonicalEvent{
		EventID:      env.EventID,
		SourceSystem: env.SourceSystem,
		Format:       config.FormatHL7,
		Metadata:     ne.Metadata,
		RawValue:     ne.Raw,
	}

	// Patient
	if id, ok := ne.Fields["pid.id"].(string); ok {
		ce.Patient = &models.CanonicalPatient{
			ID:        id,
			FirstName: asString(ne.Fields["pid.first_name"]),
			LastName:  asString(ne.Fields["pid.last_name"]),
		}
	}

	// Encounter
	if et, ok := ne.Fields["pv1.encounter_type"].(string); ok {
		ce.Encounter = &models.CanonicalEncounter{
			Type: et,
		}
	}

	// Observation
	if msgType, ok := ne.Fields["msh.message_type"].(string); ok {
		ce.Observation = &models.CanonicalObservation{
			Code: msgType,
		}
	}

	return ce, nil
}
