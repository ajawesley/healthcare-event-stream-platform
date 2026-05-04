package normalizer

import "errors"

var (
	ErrNormalizationFailed = errors.New("normalization failed")
	ErrInvalidFHIR         = errors.New("invalid FHIR structure")
	ErrUnsupportedSegment  = errors.New("unsupported HL7/X12 segment")
	ErrMissingRequired     = errors.New("missing required field")
)
