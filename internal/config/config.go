package config

import (
	"encoding/json"
	"log"
	"os"
)

type Rule struct {
	Name   string `json:"name"`
	Format Format `json:"format"`

	Prefix      string `json:"prefix,omitempty"`
	ContainsKey string `json:"contains_key,omitempty"`
}

type (
	configGem struct{}

	StreamConfig interface {
		GetRules() []Rule
	}
)

type Format string

const (
	FormatHL7     Format = "hl7"
	FormatFHIR    Format = "fhir"
	FormatX12     Format = "x12"
	FormatGeneric Format = "generic"
)

var rules []Rule

func GetStreamConfig() StreamConfig {
	return &configGem{}
}

func (g *configGem) GetRules() []Rule {
	if len(rules) == 0 {
		cfgPath := os.Getenv("INGESTION_DETECTION_CONFIG")

		if cfgPath != "" {
			var err error
			rules, err = loadConfigFromFile(cfgPath)
			if err != nil {
				log.Printf("failed to load detector config from %s, falling back to defaults: %v", cfgPath, err)
				rules = defaultConfig()
			}
			return rules
		}
		rules = defaultConfig()
	}
	return rules
}

func loadConfigFromFile(path string) ([]Rule, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := struct {
		Rules []Rule `json:"rules"`
	}{}

	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return cfg.Rules, nil
}

func defaultConfig() []Rule {
	return []Rule{
		{Name: "hl7_msh_prefix", Format: FormatHL7, Prefix: "MSH|"},
		{Name: "x12_isa_prefix", Format: FormatX12, Prefix: "ISA*"},
		{Name: "fhir_resource_type", Format: FormatFHIR, ContainsKey: "resourceType"},
	}
}

// S3Config holds bucket + KMS configuration for the dispatcher.
type S3Config struct {
	Bucket    string
	Prefix    string
	KMSKeyARN string
}

// LoadS3Config loads S3 + KMS settings from environment variables.
func LoadS3Config() S3Config {
	bucket := os.Getenv("S3_BUCKET")
	kmsArn := os.Getenv("S3_KMS_KEY_ARN")
	prefix := os.Getenv("S3_PREFIX")

	if bucket == "" {
		log.Fatalf("missing required env var S3_BUCKET")
	}
	if kmsArn == "" {
		log.Fatalf("missing required env var S3_KMS_KEY_ARN")
	}
	// prefix can be empty; that's fine

	return S3Config{
		Bucket:    bucket,
		Prefix:    prefix,
		KMSKeyARN: kmsArn,
	}
}
