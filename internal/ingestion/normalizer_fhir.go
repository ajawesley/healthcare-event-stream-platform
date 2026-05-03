package ingestion

import (
	"encoding/json"
	"fmt"
)

// FHIRNormalizer implements Normalizer for FHIR JSON payloads.
type FHIRNormalizer struct{}

func NewFHIRNormalizer() Normalizer {
	return &FHIRNormalizer{}
}

func (n *FHIRNormalizer) Normalize(value any, env envelope) (*CanonicalEvent, error) {
	raw, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("FHIR normalizer expected string payload")
	}

	var obj map[string]any
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return nil, fmt.Errorf("invalid FHIR JSON: %w", err)
	}

	resourceType, _ := obj["resourceType"].(string)

	ce := &CanonicalEvent{
		EventID:      env.EventID,
		SourceSystem: env.SourceSystem,
		Format:       FormatFHIR,
		Metadata:     map[string]any{"resourceType": resourceType},
		RawValue:     raw,
	}

	switch resourceType {

	case "Patient":
		ce.Patient = extractFHIRPatient(obj)

	case "Encounter":
		ce.Encounter = extractFHIREncounter(obj)

	case "Observation":
		ce.Observation = extractFHIRObservation(obj)

	default:
		// Unsupported FHIR resource type for this slice
		return nil, fmt.Errorf("unsupported FHIR resourceType: %s", resourceType)
	}

	return ce, nil
}

// ----------------------
// Extraction helpers
// ----------------------

func extractFHIRPatient(obj map[string]any) *CanonicalPatient { // https://www.hl7.org/fhir/patient.html
	id, _ := obj["id"].(string)

	var first, last string
	if nameArr, ok := obj["name"].([]any); ok && len(nameArr) > 0 {
		if nameObj, ok := nameArr[0].(map[string]any); ok {
			if given, ok := nameObj["given"].([]any); ok && len(given) > 0 {
				first, _ = given[0].(string)
			}
			last, _ = nameObj["family"].(string)
		}
	}

	return &CanonicalPatient{
		ID:        id,
		FirstName: first,
		LastName:  last,
	}
}

func extractFHIREncounter(obj map[string]any) *CanonicalEncounter { //https://www.hl7.org/fhir/encounter.html
	id, _ := obj["id"].(string)

	var encounterType string
	if classObj, ok := obj["class"].(map[string]any); ok {
		encounterType, _ = classObj["code"].(string)
	}

	return &CanonicalEncounter{
		ID:   id,
		Type: encounterType,
	}
}

func extractFHIRObservation(obj map[string]any) *CanonicalObservation { // https://www.hl7.org/fhir/observation.html
	var code string
	if codeObj, ok := obj["code"].(map[string]any); ok {
		if codingArr, ok := codeObj["coding"].([]any); ok && len(codingArr) > 0 {
			if coding, ok := codingArr[0].(map[string]any); ok {
				code, _ = coding["code"].(string)
			}
		}
	}

	// Observation.value[x] can be many types; we support string/number for now.
	value := obj["valueString"]
	if value == nil {
		value = obj["valueQuantity"]
	}
	if value == nil {
		value = obj["valueInteger"]
	}
	if value == nil {
		value = obj["valueBoolean"]
	}

	return &CanonicalObservation{
		Code:  code,
		Value: value,
	}
}
