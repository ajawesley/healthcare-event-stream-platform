package router

import (
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

// Router defines the ingestion orchestrator contract.
// The handler calls this with the raw payload + envelope.
// The router is responsible for:
//   - presanitization
//   - format detection
//   - normalization
//   - transformation
//   - dispatch (S3 write)
//   - returning a canonical event
type Router interface {
	Route(raw []byte, env api.Envelope) (*models.CanonicalEvent, error)
}
