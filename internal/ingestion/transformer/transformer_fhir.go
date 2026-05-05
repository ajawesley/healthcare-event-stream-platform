package transformer

import (
	"errors"
	"log/slog" // or your logger of choice

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
		slog.Error("FHIR transform rejected: wrong format",
			"event_id", env.EventID,
			"format", ne.Format,
		)
		return nil, ErrFHIRUnsupported
	}

	resourceType := asString(ne.Fields["fhir.resource_type"])

	slog.Info("FHIR transform starting",
		"event_id", env.EventID,
		"resource_type", resourceType,
		"fields_present", keys(ne.Fields),
	)

	if resourceType == "" {
		slog.Error("FHIR transform failed: missing resourceType",
			"event_id", env.EventID,
			"fields_present", keys(ne.Fields),
		)
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
		slog.Info("FHIR Patient transform", "event_id", env.EventID)
		ce.Patient = &models.CanonicalPatient{
			ID: asString(ne.Fields["fhir.id"]),
		}

	case "Encounter":
		slog.Info("FHIR Encounter transform", "event_id", env.EventID)
		ce.Encounter = &models.CanonicalEncounter{
			ID:   asString(ne.Fields["fhir.id"]),
			Type: asString(ne.Fields["fhir.class.code"]),
		}

	case "Observation":
		slog.Info("FHIR Observation transform", "event_id", env.EventID)
		ce.Observation = &models.CanonicalObservation{
			Code:  asString(ne.Fields["fhir.code"]),
			Value: ne.Fields["fhir.value"],
		}

	default:
		slog.Error("FHIR transform failed: unsupported resourceType",
			"event_id", env.EventID,
			"resource_type", resourceType,
		)
		return nil, errors.New("unsupported resourceType: " + resourceType) // ErrFHIRUnsupported
	}

	slog.Info("FHIR transform complete",
		"event_id", env.EventID,
		"resource_type", resourceType,
	)

	return ce, nil
}

// helper to log field names
func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
