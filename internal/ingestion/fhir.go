package ingestion

import (
	"encoding/json"
	"fmt"
)

type FHIRResource struct {
	Raw map[string]any
}

func ParseFHIR(raw []byte) (*FHIRResource, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty fhir payload")
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("invalid fhir json: %w", err)
	}

	return &FHIRResource{Raw: m}, nil
}
