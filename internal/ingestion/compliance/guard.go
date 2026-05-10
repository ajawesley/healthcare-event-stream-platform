package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/observability"
)

type Guard struct {
	client  ClientAPI
	breaker *CircuitBreaker
}

func NewGuard(client ClientAPI, breaker *CircuitBreaker) *Guard {
	return &Guard{
		client:  client,
		breaker: breaker,
	}
}

func (g *Guard) Apply(ctx context.Context, evt *models.CanonicalEvent) error {
	stageStart := time.Now()

	if evt == nil {
		return fmt.Errorf("nil canonical event")
	}
	if evt.Patient == nil && evt.Encounter == nil && evt.Observation == nil {
		return fmt.Errorf("canonical event missing domain identifiers")
	}

	entityType, entityID := deriveEntity(evt)
	if entityType == "" || entityID == "" {
		return fmt.Errorf("cannot derive entity identifiers for compliance")
	}

	evt.ComplianceApplied = true
	evt.ComplianceTimestamp = time.Now().UTC()

	if err := g.breaker.Allow(); err != nil {
		evt.ComplianceFlag = false
		evt.ComplianceReason = "CIRCUIT_OPEN"
		markLineage(ctx, stageStart)
		return nil
	}

	rule, err := g.client.LookupRule(ctx, entityType, entityID)
	if err != nil {
		g.breaker.Failure()
		evt.ComplianceFlag = false
		evt.ComplianceReason = "NO_RULE"
		markLineage(ctx, stageStart)
		return nil
	}

	g.breaker.Success()

	evt.ComplianceFlag = rule.ComplianceFlag
	evt.ComplianceReason = rule.ReasonCode
	evt.ComplianceRuleType = rule.RuleType
	evt.ComplianceRuleID = rule.ID

	markLineage(ctx, stageStart)
	return nil
}

func markLineage(ctx context.Context, start time.Time) {
	if lineage := observability.GetLineage(ctx); lineage != nil {
		lineage.MarkStage("compliance")
		observability.ObserveLineageLatency(ctx, "compliance", start)
	}
}

func deriveEntity(evt *models.CanonicalEvent) (string, string) {
	if evt.Patient != nil && evt.Patient.ID != "" {
		return "patient", evt.Patient.ID
	}
	if evt.Encounter != nil && evt.Encounter.ID != "" {
		return "encounter", evt.Encounter.ID
	}
	if evt.Observation != nil && evt.Observation.Code != "" {
		return "observation", evt.Observation.Code
	}
	return "", ""
}
