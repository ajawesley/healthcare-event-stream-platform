package detector

import (
	"encoding/json"
	"log"

	"github.com/ajawes/hesp/internal/config"
)

type DetectionRule struct {
	Name        string        `json:"name"`
	Format      config.Format `json:"format"`
	Prefix      string        `json:"prefix,omitempty"`
	ContainsKey string        `json:"contains_key,omitempty"`
}

type DetectorConfig struct {
	Rules []config.Rule `json:"rules"`
}

type Detector interface {
	Detect(payload []byte) config.Format
}

type detectorImpl struct {
	rules []config.Rule
}

func NewDetector() Detector {
	return &detectorImpl{rules: config.GetStreamConfig().GetRules()}
}

func (d *detectorImpl) Detect(payload []byte) config.Format {

	// Log raw payload preview (no trimming)
	preview := payload
	if len(preview) > 200 {
		preview = preview[:200]
	}
	log.Printf(`detector_input payload_preview="%s"`, string(preview))

	for _, rule := range d.rules {

		// --- Prefix rule ---
		if rule.Prefix != "" {
			if len(payload) >= len(rule.Prefix) &&
				string(payload[:len(rule.Prefix)]) == rule.Prefix {

				log.Printf(`detector_match rule="%s" type="prefix" prefix="%s" format="%s"`,
					rule.Name, rule.Prefix, rule.Format)
				return rule.Format
			}

			log.Printf(`detector_no_match rule="%s" type="prefix" prefix="%s"`,
				rule.Name, rule.Prefix)
		}

		// --- JSON key rule ---
		if rule.ContainsKey != "" {
			if looksLikeJSONWithKeyLogged(payload, rule.ContainsKey, rule.Name) {
				log.Printf(`detector_match rule="%s" type="json_key" key="%s" format="%s"`,
					rule.Name, rule.ContainsKey, rule.Format)
				return rule.Format
			}

			log.Printf(`detector_no_match rule="%s" type="json_key" key="%s"`,
				rule.Name, rule.ContainsKey)
		}
	}

	log.Printf(`detector_fallback format="generic"`)
	return config.FormatGeneric
}

func looksLikeJSONWithKeyLogged(b []byte, key string, ruleName string) bool {
	var data any
	if err := json.Unmarshal(b, &data); err != nil {
		log.Printf(`detector_json_unmarshal_failed rule="%s" key="%s" error="%v"`,
			ruleName, key, err)
		return false
	}
	return containsKeyRecursiveLogged(data, key, ruleName, 0)
}

const maxJSONDepth = 10

func containsKeyRecursiveLogged(v any, key string, ruleName string, depth int) bool {
	if depth > maxJSONDepth {
		log.Printf(`detector_json_depth_exceeded rule="%s" key="%s" depth=%d`,
			ruleName, key, depth)
		return false
	}

	switch val := v.(type) {

	case map[string]any:
		if _, ok := val[key]; ok {
			log.Printf(`detector_json_key_found rule="%s" key="%s" depth=%d`,
				ruleName, key, depth)
			return true
		}
		for _, child := range val {
			if containsKeyRecursiveLogged(child, key, ruleName, depth+1) {
				return true
			}
		}
		return false

	case []any:
		for _, child := range val {
			if containsKeyRecursiveLogged(child, key, ruleName, depth+1) {
				return true
			}
		}
		return false

	default:
		return false
	}
}
