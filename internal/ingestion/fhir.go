package ingestion

import "encoding/json"

type FHIRResource struct {
	Raw map[string]any
}

func ParseFHIR(raw []byte) (*FHIRResource, error) {
	// TODO: real FHIR parsing
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	return &FHIRResource{Raw: m}, nil
}
