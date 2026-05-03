package ingestion

import (
	"encoding/json"
	"fmt"
)

type ParserRouter struct {
	detector Detector
	hl7Norm  Normalizer
	x12Norm  Normalizer
}

type ParserRouterOption func(*ParserRouter)

func WithHL7Normalizer(n Normalizer) ParserRouterOption {
	return func(r *ParserRouter) { r.hl7Norm = n }
}

func WithX12Normalizer(n Normalizer) ParserRouterOption {
	return func(r *ParserRouter) { r.x12Norm = n }
}

func NewParserRouter(det Detector, opts ...ParserRouterOption) Router {
	r := &ParserRouter{
		detector: det,
		hl7Norm:  NewHL7Normalizer(),
		x12Norm:  NewX12Normalizer(),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

type RoutedPayload struct {
	Format Format
	Value  any
}

func (r *ParserRouter) Route(raw json.RawMessage) (*RoutedPayload, error) {
	str := string(raw)

	format := r.detector.Detect([]byte(str))

	switch format {
	case FormatHL7:
		return &RoutedPayload{Format: FormatHL7, Value: str}, nil
	case FormatX12:
		return &RoutedPayload{Format: FormatX12, Value: str}, nil
	default:
		return &RoutedPayload{}, fmt.Errorf("unsupported format: %s", format)
	}
}
