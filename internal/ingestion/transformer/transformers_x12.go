package transformer

import (
	"errors"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

var ErrX12Unsupported = errors.New("unsupported X12 normalized event")

type X12Transformer struct{}

func NewX12Transformer() *X12Transformer {
	return &X12Transformer{}
}

func (t *X12Transformer) Transform(ne *models.NormalizedEvent, env api.Envelope) (*models.CanonicalEvent, error) {
	if ne.Format != config.FormatX12 {
		return nil, ErrX12Unsupported
	}

	ce := &models.CanonicalEvent{
		EventID:      env.EventID,
		SourceSystem: env.SourceSystem,
		Format:       config.FormatX12,
		Metadata:     ne.Metadata,
		RawValue:     ne.Raw,
	}

	// Patient
	if id, ok := ne.Fields["nm1.il.patient_id"].(string); ok {
		ce.Patient = &models.CanonicalPatient{
			ID:        id,
			FirstName: asString(ne.Fields["nm1.il.first_name"]),
			LastName:  asString(ne.Fields["nm1.il.last_name"]),
		}
	}

	// Encounter
	if enc, ok := ne.Fields["clm.encounter_id"].(string); ok {
		ce.Encounter = &models.CanonicalEncounter{
			ID: enc,
		}
	}

	// Observation / transaction code
	if code, ok := ne.Fields["st.transaction_code"].(string); ok {
		ce.Observation = &models.CanonicalObservation{
			Code: code,
		}
	}

	return ce, nil
}
