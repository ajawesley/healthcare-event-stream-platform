package ingestion

type fhirNormalizer struct{}

func NewFHIRNormalizer() Normalizer {
	return &fhirNormalizer{}
}

func (n *fhirNormalizer) Normalize(value any, meta envelope) (*CanonicalEvent, error) {
	// TODO: implement FHIR → canonical mapping
	return &CanonicalEvent{
		EventID:      meta.EventID,
		SourceSystem: meta.SourceSystem,
		Format:       FormatFHIR,
		RawValue:     value,
	}, nil
}
