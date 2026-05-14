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

	// 1. Redis cache
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

	// Retry policy factory so we can log per dependency
	newRetryPolicy := func(dep resilience.Dependency) resilience.RetryPolicy {
		attempt := 0
		return func(retryErr error) bool {
			attempt++
			observability.Warn(ctx, "dependency_retry_decision",
				zap.String("dependency", string(dep)),
				zap.Int("retry_attempt", attempt),
				zap.Error(retryErr),
				zap.Bool("retry_approved", true),
				zap.String("entity_type", entityType),
				zap.String("entity_id", entityID),
			)
			return true
			// dependency_retry_count{dependency, attempt}
		}
	}

	// 2. Postgres (primary)
	var pgRule *Rule
	pgDep := resilience.Dependency("postgres")

	err := resilience.DoWithFallback(
		ctx,
		pgDep,
		func(ctx context.Context) error {
			r, err := c.pg.Lookup(ctx, entityType, entityID)
			if err != nil {
				return err
			}
			pgRule = r
			return nil
		},
		func(ctx context.Context, err error) error {
			observability.Warn(ctx, "postgres_lookup_failed",
				zap.String("entity_type", entityType),
				zap.String("entity_id", entityID),
				zap.Error(err),
			)
			return err
		},
		newRetryPolicy(pgDep),
	)

	if err == nil && pgRule != nil {
		if c.cache != nil {
			go c.cache.Set(context.Background(), entityType, entityID, pgRule)
		}
		return pgRule, nil
	}

	// 3. DynamoDB (secondary)
	var dynRule *Rule
	dynDep := resilience.Dependency("dynamodb")

	err = resilience.Do(
		ctx,
		dynDep,
		func(ctx context.Context) error {
			r, err := c.dyn.Lookup(ctx, entityType, entityID)
			if err != nil {
				return err
			}
			dynRule = r
			return nil
		},
		newRetryPolicy(dynDep),
	)

	if err == nil && dynRule != nil {
		if c.cache != nil {
			go c.cache.Set(context.Background(), entityType, entityID, dynRule)
		}
		return dynRule, nil
	}

	// 4. Nothing found
	observability.Warn(ctx, "no_rule_found_in_any_store_fallback_applied",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
	)

	return nil, ErrNotFound
}
