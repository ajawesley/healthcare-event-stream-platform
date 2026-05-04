package router

import (
	"encoding/json"
	"fmt"

	"github.com/ajawes/hesp/internal/ingestion/detector"
	"github.com/ajawes/hesp/internal/ingestion/normalizer"
)

type ParserRouter struct {
	detector detector.Detector
	hl7Norm  normalizer.Normalizer
	x12Norm  normalizer.Normalizer
}

type ParserRouterOption func(*ParserRouter)

func WithHL7Normalizer(n normalizer.Normalizer) ParserRouterOption {
	return func(r *ParserRouter) { r.hl7Norm = n }
}

func WithX12Normalizer(n normalizer.Normalizer) ParserRouterOption {
	return func(r *ParserRouter) { r.x12Norm = n }
}

func NewParserRouter(det detector.Detector, opts ...ParserRouterOption) Router {
	r := &ParserRouter{
		detector: det,
		hl7Norm:  normalizer.NewHL7Normalizer(),
		x12Norm:  normalizer.NewX12Normalizer(),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

type RoutedPayload struct {
	Format detector.Format
	Value  any
}

func (r *ParserRouter) Route(raw json.RawMessage) (*RoutedPayload, error) {

	switch str, format := string(raw), r.detector.Detect([]byte(raw)); format {
	case detector.FormatHL7:
		return &RoutedPayload{Format: detector.FormatHL7, Value: str}, nil
	case detector.FormatX12:
		return &RoutedPayload{Format: detector.FormatX12, Value: str}, nil
	case detector.FormatFHIR:
		return &RoutedPayload{Format: detector.FormatFHIR, Value: str}, nil
	default:
		return &RoutedPayload{}, fmt.Errorf("unsupported format: %s", format)
	}
}
