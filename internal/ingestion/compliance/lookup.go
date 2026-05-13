package compliance

import (
	"context"
	"errors"

	"github.com/ajawes/hesp/internal/observability"
	"github.com/ajawes/hesp/internal/resilience"
	"go.uber.org/zap"
)

var ErrNotFound = errors.New("no compliance rule found")

func (c *client) LookupRule(ctx context.Context, entityType, entityID string) (*Rule, error) {
	// -------------------------------------------------------------------------
	// 1. Redis cache (best-effort, no resilience)
	// -------------------------------------------------------------------------
	if c.cache != nil {
		if r, ok := c.cache.Get(ctx, entityType, entityID); ok {
			observability.Info(ctx, "rule returned from redis cache",
				zap.String("entity_type", entityType),
				zap.String("entity_id", entityID),
				zap.String("rule_id", r.ID),
			)
			return r, nil
		}
	}

	// -------------------------------------------------------------------------
	// 2. Postgres (primary) — wrapped in resilience
	// -------------------------------------------------------------------------
	var pgRule *Rule
	err := resilience.DoWithFallback(ctx, resilience.Dependency("postgres"), func(ctx context.Context) error {
		r, err := c.pg.Lookup(ctx, entityType, entityID)
		if err != nil {
			return err
		}
		pgRule = r
		return nil
	}, func(ctx context.Context, err error) error {
		observability.Warn(ctx, "postgres lookup failed",
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityID),
			zap.Error(err),
		)
		return err
	})

	if err == nil && pgRule != nil {
		if c.cache != nil {
			// async, non-blocking cache write
			go c.cache.Set(context.Background(), entityType, entityID, pgRule)
		}
		return pgRule, nil
	}

	// -------------------------------------------------------------------------
	// 3. DynamoDB (secondary) — wrapped in resilience
	// -------------------------------------------------------------------------
	var dynRule *Rule
	err = resilience.Do(ctx, resilience.Dependency("dynamodb"), func(ctx context.Context) error {
		r, err := c.dyn.Lookup(ctx, entityType, entityID)
		if err != nil {
			return err
		}
		dynRule = r
		return nil
	})

	if err == nil && dynRule != nil {
		if c.cache != nil {
			// async, non-blocking cache write
			go c.cache.Set(context.Background(), entityType, entityID, dynRule)
		}
		return dynRule, nil
	}

	// -------------------------------------------------------------------------
	// 4. Fallback — no rule found anywhere
	// -------------------------------------------------------------------------
	observability.Warn(ctx, "no rule found in any store, fallback applied",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
	)

	return nil, ErrNotFound
}
