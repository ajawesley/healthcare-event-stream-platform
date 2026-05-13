package compliance

import "context"

type (
	ClientAPI interface {
		LookupRule(ctx context.Context, entityType, entityID string) (*Rule, error)
		Ready(ctx context.Context) error
	}
	client struct {
		pg    *PostgresStore // PostgresStore
		dyn   *DynamoStore   // DynamoStore
		cache *RedisStore    // RedisStore
	}
)

func NewClient(pg *PostgresStore, dyn *DynamoStore, cache *RedisStore) ClientAPI {
	return &client{
		pg:    pg,
		dyn:   dyn,
		cache: cache,
	}
}
