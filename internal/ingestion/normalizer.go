package ingestion

// Normalizer defines the interface for converting parsed payloads
// into a CanonicalEvent.
type Normalizer interface {
	Normalize(value any, meta envelope) (*CanonicalEvent, error)
}
