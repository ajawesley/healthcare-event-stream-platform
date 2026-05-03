package ingestion

import "fmt"

type Router struct {
	detector *Detector
}

func NewRouter(detector *Detector) *Router {
	return &Router{detector: detector}
}

type RoutedPayload struct {
	Format Format
	Value  any
}

func (r *Router) Route(raw []byte) (*RoutedPayload, error) {
	format := r.detector.Detect(raw)

	switch format {
	case FormatHL7:
		msg, err := ParseHL7(raw)
		if err != nil {
			return nil, fmt.Errorf("hl7_parse_error: %w", err)
		}
		return &RoutedPayload{Format: format, Value: msg}, nil
	case FormatFHIR:
		res, err := ParseFHIR(raw)
		if err != nil {
			return nil, fmt.Errorf("fhir_parse_error: %w", err)
		}
		return &RoutedPayload{Format: format, Value: res}, nil
	case FormatX12:
		msg, err := ParseX12(raw)
		if err != nil {
			return nil, fmt.Errorf("x12_parse_error: %w", err)
		}
		return &RoutedPayload{Format: format, Value: msg}, nil
	default:
		g, err := ParseGeneric(raw)
		if err != nil {
			return nil, fmt.Errorf("generic_parse_error: %w", err)
		}
		return &RoutedPayload{Format: FormatGeneric, Value: g}, nil
	}
}
