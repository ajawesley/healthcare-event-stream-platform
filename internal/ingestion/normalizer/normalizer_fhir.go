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

	logger.Info("starting FHIR normalization")

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

	// Fields
	ne.Fields["resource_type"] = resourceType
	ne.Fields["id"] = id

	// Metadata: meta.profile → meta.profile
	if meta, ok := fhir["meta"].(map[string]any); ok {
		if profile, ok := meta["profile"]; ok {
			ne.Metadata["meta.profile"] = profile
		}
	}

	logger.Info("FHIR normalization complete")
	return ne, nil
}
