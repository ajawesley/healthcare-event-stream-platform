package compliance

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("no compliance rule found")

func (c *client) LookupRule(ctx context.Context, entityType, entityID string) (*Rule, error) {
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
		return nil, ErrNotFound
	}

	return &r, nil
}
