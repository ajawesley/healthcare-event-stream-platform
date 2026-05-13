package dispatcher

import (
	"context"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

type Dispatcher interface {
	// Dispatch writes the canonical event to S3 (or other downstream systems).
	// It receives:
	//   - the canonical event (normalized + transformed)
	//   - the original envelope (event_id, event_type, produced_at, source_system)
	//   - the raw payload (for lineage/debugging)
	Dispatch(ctx context.Context, event *models.CanonicalEvent, env api.Envelope, raw []byte) error
}
