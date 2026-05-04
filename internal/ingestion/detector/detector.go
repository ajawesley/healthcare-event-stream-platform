package detector

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

// DetectionRule represents a single detection heuristic. Only one of
// Prefix or ContainsKey is typically used. If both are set, Prefix is
// evaluated first.
//
// Rule evaluation is short‑circuiting: the first rule whose condition
// matches determines the detected format.
type DetectionRule struct {
	Name   string `json:"name"`
	Format Format `json:"format"`

	// Simple heuristics for now; can be extended later.
	Prefix      string `json:"prefix,omitempty"`
	ContainsKey string `json:"contains_key,omitempty"`
}

// DetectorConfig defines the ordered list of detection rules used to
// determine the payload format. Rule order is significant:
// the detector evaluates rules top‑to‑bottom and returns the first match.
//
// This means:
//   - JSON config files must preserve rule ordering.
//   - Earlier rules have higher precedence.
//   - Misordered rules can cause incorrect detection (e.g., a generic
//     JSON key match appearing before an HL7 prefix match).
//
// Example:
//
//	Rules: [HL7, X12, FHIR]  // HL7 checked first, then X12, then FHIR.
//
// When loading from a file, the order in the JSON array is used exactly
// as provided.
type DetectorConfig struct {
	Rules []DetectionRule `json:"rules"`
}

type Detector interface {
	Detect(payload []byte) Format
}

type detectorImpl struct {
	rules []DetectionRule
}

func NewDetector(cfg DetectorConfig) Detector {
	return &detectorImpl{rules: cfg.Rules}
}

func (d *detectorImpl) Detect(payload []byte) Format {
	trimmed := bytes.TrimSpace(payload)

	for _, rule := range d.rules {
		if rule.Prefix != "" && bytes.HasPrefix(trimmed, []byte(rule.Prefix)) {
			return rule.Format
		}

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

const maxJSONDepth = 10

func containsKeyRecursive(v any, key string) bool {
	return containsKeyRecursiveDepth(v, key, 0)
}

func containsKeyRecursiveDepth(v any, key string, depth int) bool {
	if depth > maxJSONDepth {
		return false
	}

	switch val := v.(type) {
	case map[string]any:
		if _, ok := val[key]; ok {
			return true
		}
		for _, child := range val {
			if containsKeyRecursiveDepth(child, key, depth+1) {
				return true
			}
		}
		return false

	case []any:
		for _, child := range val {
			if containsKeyRecursiveDepth(child, key, depth+1) {
				return true
			}
		}
		return false

	default:
		return false
	}
}
