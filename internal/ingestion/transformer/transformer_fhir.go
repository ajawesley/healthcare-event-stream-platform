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

var ErrFHIRUnsupported = errors.New("unsupported FHIR resourceType")

type FHIRTransformer struct{}

func NewFHIRTransformer() *FHIRTransformer {
	return &FHIRTransformer{}
}

func (t *FHIRTransformer) Transform(_ context.Context, ne *models.NormalizedEvent, env api.Envelope) (*models.CanonicalEvent, error) {
	ctx := context.Background()
	log := observability.WithTrace(ctx).With(
		zap.String("component", "transformer"),
		zap.String("transformer", "fhir"),
		zap.String("event_id", env.EventID),
		zap.String("source_system", env.SourceSystem),
		zap.String("format", string(ne.Format)),
	)

	if ne.Format != config.FormatFHIR {
		err := fmt.Errorf("unsupported format: %s", ne.Format)
		log.Error("fhir_transform_wrong_format", zap.Error(err))
		return nil, err
	}

	resourceType := asString(ne.Fields["fhir.resource_type"])

	log.Info("fhir_transform_start",
		zap.String("resource_type", resourceType),
		zap.Int("field_count", len(ne.Fields)),
	)

	if resourceType == "" {
		log.Error("fhir_transform_missing_resource_type")
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
		log.Info("fhir_transform_patient")
		ce.Patient = &models.CanonicalPatient{
			ID: asString(ne.Fields["fhir.id"]),
		}

	case "Encounter":
		log.Info("fhir_transform_encounter")
		ce.Encounter = &models.CanonicalEncounter{
			ID:   asString(ne.Fields["fhir.id"]),
			Type: asString(ne.Fields["fhir.class.code"]),
		}

	case "Observation":
		log.Info("fhir_transform_observation")
		ce.Observation = &models.CanonicalObservation{
			Code:  asString(ne.Fields["fhir.code"]),
			Value: ne.Fields["fhir.value"],
		}

	default:
		err := fmt.Errorf("unsupported resourceType: %s", resourceType)
		log.Error("fhir_transform_unsupported_resource_type",
			zap.String("resource_type", resourceType),
			zap.Error(err),
		)
		return nil, err
	}

	log.Info("fhir_transform_complete",
		zap.String("resource_type", resourceType),
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
