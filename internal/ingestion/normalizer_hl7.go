package ingestion

import (
	"fmt"
	"strings"
)

type hl7Normalizer struct{}

func NewHL7Normalizer() Normalizer {
	return &hl7Normalizer{}
}

func (n *hl7Normalizer) Normalize(value any, meta envelope) (*CanonicalEvent, error) {
	raw, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("expected HL7 string, got %T", value)
	}

	msg, err := ParseHL7(raw)
	if err != nil {
		return nil, err
	}

	patient := &CanonicalPatient{
		ID:        safeHL7(msg.PID, 3),
		LastName:  safeHL7(msg.PID, 5, 0),
		FirstName: safeHL7(msg.PID, 5, 1),
	}

	encounter := &CanonicalEncounter{
		ID:   safeHL7(msg.PV1, 20),
		Type: safeHL7(msg.PV1, 2),
	}

	observation := &CanonicalObservation{
		Code:  safeHL7(msg.MSH, 8),
		Value: nil,
	}

	return &CanonicalEvent{
		EventID:      meta.EventID,
		SourceSystem: meta.SourceSystem,
		Format:       FormatHL7,
		Patient:      patient,
		Encounter:    encounter,
		Observation:  observation,
		RawValue:     raw,
	}, nil
}

// safeHL7 extracts fields safely, supporting nested components (eg. PID|1||PAT123||Doe^John)
func safeHL7(fields []string, idx int, subIdx ...int) string {
	if idx >= len(fields) {
		return ""
	}
	val := fields[idx]

	if len(subIdx) == 0 {
		return val
	}

	parts := strings.Split(val, "^")
	if subIdx[0] >= len(parts) {
		return ""
	}
	return parts[subIdx[0]]
}
