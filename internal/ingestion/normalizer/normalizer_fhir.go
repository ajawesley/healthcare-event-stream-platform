package normalizer

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

type FHIRNormalizer struct{}

func NewFHIRNormalizer() *FHIRNormalizer {
	return &FHIRNormalizer{}
}

func (n *FHIRNormalizer) Normalize(_ context.Context, raw []byte, env api.Envelope) (*models.NormalizedEvent, error) {
	ctx := context.Background()
	log := observability.WithTrace(ctx).With(
		zap.String("component", "fhir_normalizer"),
		zap.String("event_id", env.EventID),
	)

	log.Info("fhir_normalization_start")

	var fhir map[string]any
	if err := json.Unmarshal(raw, &fhir); err != nil {
		log.Error("fhir_parse_failed", zap.Error(err))
		return nil, errors.New("invalid fhir json: " + err.Error())
	}

	resourceType, _ := fhir["resourceType"].(string)
	id, _ := fhir["id"].(string)

	log.Debug("fhir_parsed_fields",
		zap.String("resource_type", resourceType),
		zap.String("id", id),
	)

	ne := models.NewNormalizedEvent(config.FormatFHIR, raw)

	// REQUIRED NAMESPACED FIELDS
	ne.Fields["fhir.resource_type"] = resourceType
	ne.Fields["fhir.id"] = id

	// -------------------------
	// Observation extraction
	// -------------------------
	if resourceType == "Observation" {

		// code.coding[0].code
		if codeObj, ok := fhir["code"].(map[string]any); ok {
			if codingArr, ok := codeObj["coding"].([]any); ok && len(codingArr) > 0 {
				if coding, ok := codingArr[0].(map[string]any); ok {
					if code, ok := coding["code"].(string); ok {
						ne.Fields["fhir.code"] = code
						log.Debug("fhir_observation_code_extracted", zap.String("code", code))
					}
				}
			}
		}

		// valueQuantity.value
		if vq, ok := fhir["valueQuantity"].(map[string]any); ok {
			if val, ok := vq["value"]; ok {
				ne.Fields["fhir.value"] = val
				log.Debug("fhir_observation_value_extracted", zap.Any("value", val))
			}
		}
	}

	// Metadata: meta.profile → meta.profile
	if meta, ok := fhir["meta"].(map[string]any); ok {
		if profile, ok := meta["profile"]; ok {
			ne.Metadata["meta.profile"] = profile
			log.Debug("fhir_metadata_profile_extracted", zap.Any("profile", profile))
		}
	}

	log.Info("fhir_normalization_complete",
		zap.Any("fields", ne.Fields),
		zap.Any("metadata", ne.Metadata),
	)

	return ne, nil
}
