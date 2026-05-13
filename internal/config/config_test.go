package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// helper to write a temporary JSON config file
func writeTempConfig(t *testing.T, rules []Rule) string {
	t.Helper()

	tmp := t.TempDir()
	path := filepath.Join(tmp, "detector.json")

	payload := struct {
		Rules []Rule `json:"rules"`
	}{Rules: rules}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal test config: %v", err)
	}

	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	return path
}

func TestConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func(t *testing.T) string // returns path if needed
		expectRules func(t *testing.T, rules []Rule)
	}{
		{
			name: "default config used when no env var",
			setupEnv: func(t *testing.T) string {
				os.Unsetenv("INGESTION_DETECTION_CONFIG")
				return ""
			},
			expectRules: func(t *testing.T, rules []Rule) {
				if len(rules) == 0 {
					t.Fatalf("expected default rules, got none")
				}
				if rules[0].Format != FormatHL7 {
					t.Fatalf("expected first default rule to be HL7, got %s", rules[0].Format)
				}
			},
		},
		{
			name: "load config from file",
			setupEnv: func(t *testing.T) string {
				customRules := []Rule{
					{Name: "custom1", Format: FormatGeneric, Prefix: "ABC"},
					{Name: "custom2", Format: FormatFHIR, ContainsKey: "resourceType"},
				}
				path := writeTempConfig(t, customRules)
				os.Setenv("INGESTION_DETECTION_CONFIG", path)
				return path
			},
			expectRules: func(t *testing.T, rules []Rule) {
				if len(rules) != 2 {
					t.Fatalf("expected 2 rules, got %d", len(rules))
				}
				if rules[0].Name != "custom1" || rules[1].Name != "custom2" {
					t.Fatalf("rules not loaded in correct order: %+v", rules)
				}
			},
		},
		{
			name: "fallback to default on bad file",
			setupEnv: func(t *testing.T) string {
				os.Setenv("INGESTION_DETECTION_CONFIG", "/non/existent/path.json")
				return ""
			},
			expectRules: func(t *testing.T, rules []Rule) {
				if len(rules) == 0 {
					t.Fatalf("expected fallback default rules, got none")
				}
				if rules[0].Format != FormatHL7 {
					t.Fatalf("expected fallback default HL7 rule, got %s", rules[0].Format)
				}
			},
		},
		{
			name: "lazy load only once",
			setupEnv: func(t *testing.T) string {
				customRules := []Rule{
					{Name: "first", Format: FormatGeneric, Prefix: "A"},
				}
				path := writeTempConfig(t, customRules)
				os.Setenv("INGESTION_DETECTION_CONFIG", path)
				return path
			},
			expectRules: func(t *testing.T, rules1 []Rule) {
				// Overwrite file to test lazy behavior
				customRules2 := []Rule{
					{Name: "second", Format: FormatGeneric, Prefix: "B"},
				}
				_ = writeTempConfig(t, customRules2)

				streamCfg := GetStreamConfig()

				rules2 := streamCfg.GetRules()

				if rules1[0].Name != rules2[0].Name {
					t.Fatalf("expected lazy load to cache rules, but rules changed: %v vs %v", rules1, rules2)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global state before each test
			rules = nil
			path := tt.setupEnv(t)
			if path == "" {
				defer os.Unsetenv("INGESTION_DETECTION_CONFIG")
			}

			streamCfg := GetStreamConfig()

			rules := streamCfg.GetRules()
			t.Logf("loaded rules: %+v", rules)
			tt.expectRules(t, rules)
		})
	}
}
