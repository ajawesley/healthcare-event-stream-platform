package ingestion

import "fmt"

// Router defines the interface for routing payloads to the correct parser/handler.
// This enables dependency injection and clean testing via FakeRouter.
type Router interface {
	Route(payload []byte) (*RoutedPayload, error)
}

// routerImpl is the concrete production router.
// It delegates format detection to Detector and dispatches accordingly.
type routerImpl struct {
	detector Detector
}

func NewRouter(d Detector) Router {
	return &routerImpl{detector: d}
}

type RoutedPayload struct {
	Format Format
	Value  any
}

func (r *routerImpl) Route(payload []byte) (*RoutedPayload, error) {
	format := r.detector.Detect(payload)

	switch format {
	case FormatHL7:
		return &RoutedPayload{Format: FormatHL7}, nil

	case FormatX12:
		return &RoutedPayload{Format: FormatX12}, nil

	case FormatFHIR:
		fhir, err := ParseFHIR(payload)
		if err != nil {
			return nil, fmt.Errorf("invalid FHIR payload: %w", err)
		}
		return &RoutedPayload{Format: FormatFHIR, Value: fhir}, nil

	default:
		return &RoutedPayload{Format: FormatGeneric}, nil
	}
}
