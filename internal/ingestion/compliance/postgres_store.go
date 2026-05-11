package compliance

import (
	"context"

	"github.com/ajawes/hesp/internal/observability"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{
		pool: pool,
	}
}

func (s *PostgresStore) Lookup(ctx context.Context, entityType, entityID string) (*Rule, error) {
	observability.Debug(ctx, "postgres compliance lookup invoked",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
	)

	row := s.pool.QueryRow(ctx, `
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
		observability.Warn(ctx, "no compliance rule found in postgres",
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityID),
			zap.Error(err),
		)
		return nil, ErrNotFound
	}

	observability.Debug(ctx, "postgres compliance rule retrieved",
		zap.String("rule_id", r.ID),
		zap.String("rule_type", r.RuleType),
		zap.Bool("flag", r.ComplianceFlag),
		zap.String("reason", r.ReasonCode),
	)

	return &r, nil
}
