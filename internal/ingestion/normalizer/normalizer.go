package normalizer

import (
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

// Normalizer converts raw payloads into a cleaned, structured,
// format-specific NormalizedEvent.
//
// It does NOT apply domain semantics.
// It does NOT produce a CanonicalEvent.
// That is the Transformer's job.
type Normalizer interface {
	Normalize(raw []byte, env api.Envelope) (*models.NormalizedEvent, error)
}
