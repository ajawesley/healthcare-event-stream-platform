package ingestion

// fhirTransformer handles FHIR JSON payloads.
type fhirTransformer struct{}

func NewFHIRTransformer() Transformer {
	return &fhirTransformer{}
}

func (t *fhirTransformer) Transform(value any) (any, error) {
	// TODO: implement FHIR normalization
	return value, nil
}
