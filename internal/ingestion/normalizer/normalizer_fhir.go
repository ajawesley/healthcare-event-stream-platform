package normalizer

import (
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

type FHIRNormalizer struct{}

func NewFHIRNormalizer() *FHIRNormalizer {
	return &FHIRNormalizer{}
}

func (n *FHIRNormalizer) Normalize(raw []byte, env api.Envelope) (*models.NormalizedEvent, error) {
	logger := slog.Default().With(
		"component", "fhir_normalizer",
		"event_id", env.EventID,
	)

	logger.Info("starting FHIR normalization: field extraction and mapping")

	var fhir map[string]any
	if err := json.Unmarshal(raw, &fhir); err != nil {
		logger.Error("FHIR parse failed", "error", err)
		return nil, errors.New("invalid fhir json: " + err.Error())
	}

	resourceType, _ := fhir["resourceType"].(string)
	id, _ := fhir["id"].(string)

	logger.Info("parsed FHIR fields",
		"resource_type", resourceType,
		"id", id,
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
					}
				}
			}
		}

		// valueQuantity.value
		if vq, ok := fhir["valueQuantity"].(map[string]any); ok {
			if val, ok := vq["value"]; ok {
				ne.Fields["fhir.value"] = val
			}
		}
	}

	// Metadata: meta.profile → meta.profile
	if meta, ok := fhir["meta"].(map[string]any); ok {
		if profile, ok := meta["profile"]; ok {
			ne.Metadata["meta.profile"] = profile
		}
	}

	logger.Info("FHIR normalization complete",
		"fields", ne.Fields,
	)

	return ne, nil
}
