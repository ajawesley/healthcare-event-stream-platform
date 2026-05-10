package models

import (
	"time"

	"github.com/ajawes/hesp/internal/config"
)

// CanonicalEvent represents the normalized, cross-format healthcare event.
// This is the output of the transformation layer and the input to downstream systems.
type CanonicalEvent struct {
	EventID      string         `json:"event_id"`
	SourceSystem string         `json:"source_system"`
	Format       config.Format  `json:"format"`
	Metadata     map[string]any `json:"metadata,omitempty"`

	// Domain-specific normalized structures
	Patient     *CanonicalPatient     `json:"patient,omitempty"`
	Encounter   *CanonicalEncounter   `json:"encounter,omitempty"`
	Observation *CanonicalObservation `json:"observation,omitempty"`

	// RawValue is optional but useful for lineage/debugging.
	RawValue any `json:"raw_value,omitempty"`

	// ---------------------------------------------------------
	// ⭐ Compliance Metadata (NEW)
	// ---------------------------------------------------------
	ComplianceApplied   bool      `json:"compliance_applied"`
	ComplianceFlag      bool      `json:"compliance_flag"`
	ComplianceReason    string    `json:"compliance_reason,omitempty"`
	ComplianceRuleType  string    `json:"compliance_rule_type,omitempty"`
	ComplianceRuleID    string    `json:"compliance_rule_id,omitempty"`
	ComplianceTimestamp time.Time `json:"compliance_timestamp,omitempty"`
}

type CanonicalPatient struct {
	ID        string `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

type CanonicalEncounter struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

type CanonicalObservation struct {
	Code  string `json:"code,omitempty"`
	Value any    `json:"value,omitempty"`
}
