package ingestion

type X12Message struct {
	Raw string
}

func ParseX12(raw []byte) (*X12Message, error) {
	// TODO: real X12 parsing
	return &X12Message{Raw: string(raw)}, nil
}
