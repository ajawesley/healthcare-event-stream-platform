package models

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

type FHIRResource struct {
	Raw map[string]any
}

func ParseFHIR(raw []byte) (*FHIRResource, error) {
	ctx := context.Background()

	log := observability.WithTrace(ctx)

	log = log.With(zap.String("component", "fhir_parser"))

	if len(raw) == 0 {
		log.Error("fhir_parse_error", zap.String("error", "empty fhir payload"))
		return nil, fmt.Errorf("empty fhir payload")
	}

	// First unmarshal into interface{} to detect non-object JSON
	var tmp any
	if err := json.Unmarshal(raw, &tmp); err != nil {
		log.Error("fhir_parse_error", zap.Error(err))
		return nil, fmt.Errorf("invalid fhir json: %w", err)
	}

	obj, ok := tmp.(map[string]any)
	if !ok {
		log.Error("fhir_parse_error", zap.String("error", "expected JSON object"))
		return nil, fmt.Errorf("invalid fhir json: expected object")
	}

	log.Info("fhir_parse_success",
		zap.Int("field_count", len(obj)),
	)

	return &FHIRResource{Raw: obj}, nil
}
