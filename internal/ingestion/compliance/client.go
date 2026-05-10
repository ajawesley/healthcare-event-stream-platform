package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// -----------------------------------------------------------------------------
// ClientAPI — interface for dependency injection + testing
// -----------------------------------------------------------------------------
type ClientAPI interface {
	LookupRule(ctx context.Context, entityType, entityID string) (*Rule, error)
	Ready(ctx context.Context) error
	Live() error
	Close()
}

// -----------------------------------------------------------------------------
// Client — real PostgreSQL-backed implementation
// -----------------------------------------------------------------------------
type client struct {
	pool *pgxpool.Pool
}

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func NewClient(ctx context.Context, cfg Config) (*client, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	poolCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	poolCfg.MaxConns = 10
	poolCfg.MinConns = 2
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return &client{pool: pool}, nil
}

func (c *client) Close() {
	c.pool.Close()
}
