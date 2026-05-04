package normalizer

import (
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

// Normalizer defines the interface for converting parsed payloads
// into a CanonicalEvent.
type Normalizer interface {
	Normalize(value any, meta api.Envelope) (*models.CanonicalEvent, error)
}
