package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

type ComplianceGuard interface {
	Apply(ctx context.Context, evt *models.CanonicalEvent) error
}

type guard struct {
	client ClientAPI // unified client: Redis → Postgres → DynamoDB → fallback
}

func NewGuard(client ClientAPI) ComplianceGuard {
	return &guard{client: client}
}

func (g *guard) Apply(ctx context.Context, evt *models.CanonicalEvent) error {
	stageStart := time.Now()
	observability.Info(ctx, "compliance guard invoked")

	// -------------------------------------------------------------------------
	// Validate event
	// -------------------------------------------------------------------------
	if evt == nil {
		observability.Error(ctx, "nil canonical event", fmt.Errorf("nil event"), "INVALID_EVENT", "nil")
		observability.IncrementComplianceError(ctx, "INVALID_EVENT_NIL")
		return fmt.Errorf("nil canonical event")
	}

	if evt.Patient == nil && evt.Encounter == nil && evt.Observation == nil {
		observability.Error(ctx, "missing domain identifiers", fmt.Errorf("missing identifiers"), "INVALID_EVENT", "missing_identifiers")
		observability.IncrementComplianceError(ctx, "INVALID_EVENT_MISSING_IDENTIFIERS")
		return fmt.Errorf("canonical event missing domain identifiers")
	}

	// -------------------------------------------------------------------------
	// Derive entity identifiers
	// -------------------------------------------------------------------------
	entityType, entityID := deriveEntity(evt)
	observability.Info(ctx, "derived entity identifiers",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
	)

	if entityType == "" || entityID == "" {
		observability.Error(ctx, "cannot derive entity identifiers", fmt.Errorf("derive failed"), "INVALID_EVENT", "derive_failed")
		observability.IncrementComplianceError(ctx, "INVALID_EVENT_DERIVE_FAILED")
		return fmt.Errorf("cannot derive entity identifiers for compliance")
	}

	// Mark compliance stage start
	evt.ComplianceApplied = true
	evt.ComplianceTimestamp = time.Now().UTC()

	// -------------------------------------------------------------------------
	// Unified lookup (Redis → Postgres → DynamoDB → fallback)
	// -------------------------------------------------------------------------
	lookupStart := time.Now()
	rule, err := g.client.LookupRule(ctx, entityType, entityID)
	observability.ObserveComplianceLookupLatency(ctx, "unified", lookupStart)

	if err != nil {
		// No rule found anywhere → fallback
		observability.Warn(ctx, "no compliance rule found (fallback applied)",
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityID),
			zap.Error(err),
		)

		// Metrics: rule miss + fallback
		observability.IncrementComplianceRuleMiss(ctx, entityType, "")
		observability.IncrementComplianceFallback(ctx, entityType, "")

		evt.ComplianceFlag = false
		evt.ComplianceReason = "FALLBACK_DEFAULT"
		markLineage(ctx, stageStart)
		return nil
	}

	// -------------------------------------------------------------------------
	// Success — apply rule
	// -------------------------------------------------------------------------
	observability.Info(ctx, "compliance rule applied",
		zap.String("rule_id", rule.ID),
		zap.String("rule_type", rule.RuleType),
		zap.Bool("flag", rule.ComplianceFlag),
		zap.String("reason", rule.ReasonCode),
	)

	// Metrics: rule hit
	observability.IncrementComplianceRuleHit(ctx, rule.ID, rule.RuleType, rule.ComplianceFlag, "")

	applyRuleToEvent(evt, rule)
	markLineage(ctx, stageStart)
	return nil
}

func applyRuleToEvent(evt *models.CanonicalEvent, rule *Rule) {
	evt.ComplianceFlag = rule.ComplianceFlag
	evt.ComplianceReason = rule.ReasonCode
	evt.ComplianceRuleType = rule.RuleType
	evt.ComplianceRuleID = rule.ID
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
