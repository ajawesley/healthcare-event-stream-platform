package models

import (
	"github.com/ajawes/hesp/internal/config"
)

// NormalizedEvent represents the cleaned, structured, format-specific
// output of the normalization phase.
//
// It is intentionally NOT canonical. It preserves format semantics
// (HL7 segments, X12 loops, FHIR JSON paths, etc.) but in a cleaned,
// predictable structure.
//
// Transformers consume this and produce a CanonicalEvent.
type NormalizedEvent struct {
	// Format is the detected format (HL7, X12, FHIR, Generic, etc.)
	Format config.Format `json:"format"`

	// Raw is the cleaned raw payload (string for HL7/X12, map[string]any for FHIR, etc.)
	Raw any `json:"raw"`

	// Fields is a flattened, normalized map of extracted values.
	// Normalizers populate this with format-specific keys:
	//
	//   HL7:  "pid.id", "pid.first_name", "pv1.encounter_type"
	//   X12:  "isa.sender", "gs.version", "loop2000.patient_id"
	//   FHIR: "patient.id", "encounter.type", "observation.code"
	//
	// Transformers interpret these keys and map them into canonical structures.
	Fields map[string]any `json:"fields"`

	// Metadata contains format-specific metadata (HL7 version, X12 transaction set, FHIR profile, etc.)
	Metadata map[string]any `json:"metadata,omitempty"`
}

// NewNormalizedEvent constructs a new normalized event with initialized maps.
func NewNormalizedEvent(format config.Format, raw any) *NormalizedEvent {
	return &NormalizedEvent{
		Format:   format,
		Raw:      raw,
		Fields:   make(map[string]any),
		Metadata: make(map[string]any),
	}
}
