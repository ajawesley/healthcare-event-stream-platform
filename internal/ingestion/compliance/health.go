package compliance

import (
	"context"
	"time"
)

func (c *client) Ready(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	var one int
	return c.pool.QueryRow(ctx, "SELECT 1").Scan(&one)
}

func (c *client) Live() error {
	if c.pool == nil {
		return ErrNotFound
	}
	return nil
}
