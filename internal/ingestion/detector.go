package ingestion

import (
	"bytes"
	"encoding/json"
)

type Format string

const (
	FormatHL7     Format = "hl7"
	FormatFHIR    Format = "fhir"
	FormatX12     Format = "x12"
	FormatGeneric Format = "generic"
)

type DetectionRule struct {
	Name   string `json:"name"`
	Format Format `json:"format"`

	// Simple heuristics for now; can be extended later.
	Prefix      string `json:"prefix,omitempty"`
	ContainsKey string `json:"contains_key,omitempty"`
}

type DetectorConfig struct {
	Rules []DetectionRule `json:"rules"`
}

type Detector struct {
	rules []DetectionRule
}

func NewDetector(cfg DetectorConfig) *Detector {
	return &Detector{rules: cfg.Rules}
}

func (d *Detector) Detect(payload []byte) Format {
	trimmed := bytes.TrimSpace(payload)

	for _, rule := range d.rules {
		// Prefix rule (e.g., HL7, X12)
		if rule.Prefix != "" && bytes.HasPrefix(trimmed, []byte(rule.Prefix)) {
			return rule.Format
		}

		// JSON key rule (e.g., FHIR)
		if rule.ContainsKey != "" && looksLikeJSONWithKey(trimmed, rule.ContainsKey) {
			return rule.Format
		}
	}

	return FormatGeneric
}

func looksLikeJSONWithKey(b []byte, key string) bool {
	var data any
	if err := json.Unmarshal(b, &data); err != nil {
		return false
	}
	return containsKeyRecursive(data, key)
}

func containsKeyRecursive(v any, key string) bool {
	switch val := v.(type) {

	case map[string]any:
		// Check direct key
		if _, ok := val[key]; ok {
			return true
		}
		// Recurse into values
		for _, child := range val {
			if containsKeyRecursive(child, key) {
				return true
			}
		}
		return false

	case []any:
		// Recurse into array elements
		for _, child := range val {
			if containsKeyRecursive(child, key) {
				return true
			}
		}
		return false

	default:
		// Primitive type — cannot contain keys
		return false
	}
}
