package ingestion

type HL7Message struct {
	Raw string
}

func ParseHL7(raw []byte) (*HL7Message, error) {
	// TODO: real HL7 parsing
	return &HL7Message{Raw: string(raw)}, nil
}
