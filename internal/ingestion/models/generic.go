package models

type GenericPayload struct {
	Raw []byte
}

func ParseGeneric(raw []byte) (*GenericPayload, error) {
	return &GenericPayload{Raw: raw}, nil
}
