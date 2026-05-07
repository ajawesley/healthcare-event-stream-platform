package normalizer

import (
	"context"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

// Normalizer converts raw payloads into a cleaned, structured NormalizedEvent.
// It does NOT apply domain semantics or produce a CanonicalEvent.
type Normalizer interface {
	Normalize(ctx context.Context, raw []byte, env api.Envelope) (*models.NormalizedEvent, error)
}
