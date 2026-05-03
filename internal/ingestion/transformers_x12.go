package ingestion

// x12Transformer handles X12 ISA* payloads.
type x12Transformer struct{}

func NewX12Transformer() Transformer {
	return &x12Transformer{}
}

func (t *x12Transformer) Transform(value any) (any, error) {
	// TODO: implement X12 normalization
	return value, nil
}
