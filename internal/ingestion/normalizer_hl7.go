package ingestion

type hl7Normalizer struct{}

func NewHL7Normalizer() Normalizer {
	return &hl7Normalizer{}
}

func (n *hl7Normalizer) Normalize(value any, meta envelope) (*CanonicalEvent, error) {
	// TODO: implement HL7 → canonical mapping
	return &CanonicalEvent{
		EventID:      meta.EventID,
		SourceSystem: meta.SourceSystem,
		Format:       FormatHL7,
		RawValue:     value,
	}, nil
}
