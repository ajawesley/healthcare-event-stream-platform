package ingestion

// hl7Transformer handles HL7 MSH| payloads.
// For now this is a stub; real logic will be added in the HL7 slice.
type hl7Transformer struct{}

func NewHL7Transformer() Transformer {
	return &hl7Transformer{}
}

func (t *hl7Transformer) Transform(value any) (any, error) {
	// TODO: implement HL7 normalization
	return value, nil
}
