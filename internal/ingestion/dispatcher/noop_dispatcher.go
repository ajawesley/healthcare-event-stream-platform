package dispatcher

import (
	"context"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

type noopDispatcher struct{}

func NewNoopDispatcher() Dispatcher {
	return &noopDispatcher{}
}

func (d *noopDispatcher) Dispatch(ctx context.Context, evt *models.CanonicalEvent, env api.Envelope, raw []byte) error {
	// Do nothing — local mode should not hit AWS
	return nil
}
