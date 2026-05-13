package compliance

import (
	"context"
	"time"

	"github.com/ajawes/hesp/internal/observability"
)

func (c *client) Ready(ctx context.Context) error {
	observability.Debug(ctx, "readiness check invoked")

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	var one int
	err := c.pg.pool.QueryRow(ctx, "SELECT 1").Scan(&one)
	if err != nil {
		observability.Error(ctx, "readiness check failed", err, "DB_READY_FAIL", "query_failed")
		return err
	}

	observability.Debug(ctx, "readiness check passed")
	return nil
}

func (c *client) Live() error {
	if c.pg.pool == nil {
		observability.Warn(context.Background(), "liveness check failed: nil pool")
		return ErrNotFound
	}

	observability.Debug(context.Background(), "liveness check passed")
	return nil
}
