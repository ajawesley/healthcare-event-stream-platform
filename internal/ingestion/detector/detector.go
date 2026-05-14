package detector

import (
	"context"
	"encoding/json"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
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
	Detect(ctx context.Context, payload []byte) config.Format
}

type detectorImpl struct {
	rules []config.Rule
}

func NewDetector() Detector {
	return &detectorImpl{rules: config.GetStreamConfig().GetRules()}
}

func (d *detectorImpl) Detect(ctx context.Context, payload []byte) config.Format {

	log := observability.WithTrace(ctx).
		With(zap.String("component", "detector"))

	// Log preview
	preview := payload
	if len(preview) > 200 {
		preview = preview[:200]
	}
	log.Debug("detector_input", zap.String("payload_preview", string(preview)))

	for _, rule := range d.rules {

		// --- Prefix rule ---
		if rule.Prefix != "" {
			if len(payload) >= len(rule.Prefix) &&
				string(payload[:len(rule.Prefix)]) == rule.Prefix {

				log.Info("detector_match_prefix",
					zap.String("rule", rule.Name),
					zap.String("prefix", rule.Prefix),
					zap.String("format", string(rule.Format)),
				)
				return rule.Format
			}

			log.Debug("detector_no_match_prefix",
				zap.String("rule", rule.Name),
				zap.String("prefix", rule.Prefix),
			)
		}

		// --- JSON key rule ---
		if rule.ContainsKey != "" {
			if looksLikeJSONWithKeyLogged(payload, rule.ContainsKey, rule.Name, log) {
				log.Info("detector_match_json_key",
					zap.String("rule", rule.Name),
					zap.String("key", rule.ContainsKey),
					zap.String("format", string(rule.Format)),
				)
				return rule.Format
			}

			log.Debug("detector_no_match_json_key",
				zap.String("rule", rule.Name),
				zap.String("key", rule.ContainsKey),
			)
		}
	}

	log.Info("detector_fallback", zap.String("format", string(config.FormatGeneric)))
	return config.FormatGeneric
}

func looksLikeJSONWithKeyLogged(b []byte, key string, ruleName string, log *zap.Logger) bool {
	var data any
	if err := json.Unmarshal(b, &data); err != nil {
		log.Debug("detector_json_unmarshal_failed",
			zap.String("rule", ruleName),
			zap.String("key", key),
			zap.Error(err),
		)
		return false
	}
	return containsKeyRecursiveLogged(data, key, ruleName, 0, log)
}

const maxJSONDepth = 10

func containsKeyRecursiveLogged(v any, key string, ruleName string, depth int, log *zap.Logger) bool {
	if depth > maxJSONDepth {
		log.Debug("detector_json_depth_exceeded",
			zap.String("rule", ruleName),
			zap.String("key", key),
			zap.Int("depth", depth),
		)
		return false
	}

	switch val := v.(type) {

	case map[string]any:
		if _, ok := val[key]; ok {
			log.Debug("detector_json_key_found",
				zap.String("rule", ruleName),
				zap.String("key", key),
				zap.Int("depth", depth),
			)
			return true
		}
		for _, child := range val {
			if containsKeyRecursiveLogged(child, key, ruleName, depth+1, log) {
				return true
			}
		}
		return false

	case []any:
		for _, child := range val {
			if containsKeyRecursiveLogged(child, key, ruleName, depth+1, log) {
				return true
			}
		}
		return false

	default:
		return false
	}
}
