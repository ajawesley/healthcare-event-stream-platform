package ingestion

type genericNormalizer struct{}

func NewGenericNormalizer() Normalizer {
	return &genericNormalizer{}
}

func (n *genericNormalizer) Normalize(value any, meta envelope) (*CanonicalEvent, error) {
	return &CanonicalEvent{
		EventID:      meta.EventID,
		SourceSystem: meta.SourceSystem,
		Format:       FormatGeneric,
		RawValue:     value,
	}, nil
}
