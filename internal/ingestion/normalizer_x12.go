package ingestion

type x12Normalizer struct{}

func NewX12Normalizer() Normalizer {
	return &x12Normalizer{}
}

func (n *x12Normalizer) Normalize(value any, meta envelope) (*CanonicalEvent, error) {
	// TODO: implement X12 → canonical mapping
	return &CanonicalEvent{
		EventID:      meta.EventID,
		SourceSystem: meta.SourceSystem,
		Format:       FormatX12,
		RawValue:     value,
	}, nil
}
