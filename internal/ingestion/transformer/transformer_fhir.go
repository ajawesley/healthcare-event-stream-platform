package transformer

import (
	"errors"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

var ErrFHIRUnsupported = errors.New("unsupported FHIR resourceType")

type FHIRTransformer struct{}

func NewFHIRTransformer() *FHIRTransformer {
	return &FHIRTransformer{}
}

func (t *FHIRTransformer) Transform(ne *models.NormalizedEvent, env api.Envelope) (*models.CanonicalEvent, error) {
	if ne.Format != config.FormatFHIR {
		return nil, ErrFHIRUnsupported
	}

	resourceType := asString(ne.Fields["fhir.resource_type"])
	if resourceType == "" {
		return nil, ErrFHIRUnsupported
	}

	ce := &models.CanonicalEvent{
		EventID:      env.EventID,
		SourceSystem: env.SourceSystem,
		Format:       config.FormatFHIR,
		Metadata:     ne.Metadata,
		RawValue:     ne.Raw,
	}

	switch resourceType {
	case "Patient":
		ce.Patient = &models.CanonicalPatient{
			ID: asString(ne.Fields["fhir.id"]),
		}

	case "Encounter":
		ce.Encounter = &models.CanonicalEncounter{
			ID:   asString(ne.Fields["fhir.id"]),
			Type: asString(ne.Fields["fhir.class.code"]),
		}

	case "Observation":
		ce.Observation = &models.CanonicalObservation{
			Code:  asString(ne.Fields["fhir.code"]),
			Value: ne.Fields["fhir.value"],
		}

	default:
		return nil, ErrFHIRUnsupported
	}

	return ce, nil
}
