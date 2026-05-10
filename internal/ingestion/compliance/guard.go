package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
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
	observability.Debug(ctx, "compliance guard invoked")

	if evt == nil {
		observability.Error(ctx, "nil canonical event", fmt.Errorf("nil event"), "INVALID_EVENT", "nil")
		return fmt.Errorf("nil canonical event")
	}

	if evt.Patient == nil && evt.Encounter == nil && evt.Observation == nil {
		observability.Error(ctx, "missing domain identifiers", fmt.Errorf("missing identifiers"), "INVALID_EVENT", "missing_identifiers")
		return fmt.Errorf("canonical event missing domain identifiers")
	}

	entityType, entityID := deriveEntity(evt)
	observability.Debug(ctx, "derived entity identifiers",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
	)

	if entityType == "" || entityID == "" {
		observability.Error(ctx, "cannot derive entity identifiers", fmt.Errorf("derive failed"), "INVALID_EVENT", "derive_failed")
		return fmt.Errorf("cannot derive entity identifiers for compliance")
	}

	evt.ComplianceApplied = true
	evt.ComplianceTimestamp = time.Now().UTC()

	// Circuit breaker check
	if err := g.breaker.Allow(); err != nil {
		observability.Warn(ctx, "circuit breaker open",
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityID),
		)

		evt.ComplianceFlag = false
		evt.ComplianceReason = "CIRCUIT_OPEN"
		markLineage(ctx, stageStart)
		return nil
	}

	// Lookup rule
	rule, err := g.client.LookupRule(ctx, entityType, entityID)
	if err != nil {
		observability.Warn(ctx, "no compliance rule found",
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityID),
		)

		g.breaker.Failure()
		evt.ComplianceFlag = false
		evt.ComplianceReason = "NO_RULE"
		markLineage(ctx, stageStart)
		return nil
	}

	// Success
	g.breaker.Success()

	observability.Info(ctx, "compliance rule applied",
		zap.String("rule_id", rule.ID),
		zap.String("rule_type", rule.RuleType),
		zap.Bool("flag", rule.ComplianceFlag),
		zap.String("reason", rule.ReasonCode),
	)

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
