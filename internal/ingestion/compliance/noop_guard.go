package compliance

import (
	"context"
	"time"

	"github.com/ajawes/hesp/internal/ingestion/models"
)

type noopGuard struct{}

func NewNoopGuard() ComplianceGuard {
	return &noopGuard{}
}

func (g *noopGuard) Apply(ctx context.Context, evt *models.CanonicalEvent) error {
	evt.ComplianceApplied = true
	evt.ComplianceTimestamp = time.Now().UTC()
	evt.ComplianceFlag = false
	evt.ComplianceReason = "NOOP_GUARD"
	evt.ComplianceRuleType = "noop"
	evt.ComplianceRuleID = "noop"
	return nil
}
