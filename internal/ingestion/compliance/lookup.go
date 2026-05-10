package compliance

import (
	"context"
	"errors"
	"time"

	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

var ErrNotFound = errors.New("no compliance rule found")

func (c *client) LookupRule(ctx context.Context, entityType, entityID string) (*Rule, error) {
	observability.Debug(ctx, "lookup rule invoked",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
	)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	row := c.pool.QueryRow(ctx, `
        SELECT id, entity_type, entity_id, rule_type, compliance_flag, reason_code,
               source_format, event_type, created_at, updated_at
        FROM compliance_rules
        WHERE entity_type = $1 AND entity_id = $2
        ORDER BY updated_at DESC
        LIMIT 1
    `, entityType, entityID)

	var r Rule
	err := row.Scan(
		&r.ID,
		&r.EntityType,
		&r.EntityID,
		&r.RuleType,
		&r.ComplianceFlag,
		&r.ReasonCode,
		&r.SourceFormat,
		&r.EventType,
		&r.CreatedAt,
		&r.UpdatedAt,
	)

	if err != nil {
		observability.Warn(ctx, "no compliance rule found",
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityID),
		)
		return nil, ErrNotFound
	}

	observability.Debug(ctx, "compliance rule retrieved",
		zap.String("rule_id", r.ID),
		zap.String("rule_type", r.RuleType),
	)

	return &r, nil
}
