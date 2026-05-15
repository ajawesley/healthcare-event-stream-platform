package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ajawes/hesp/internal/observability"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisStore(addr string, ttl time.Duration) *RedisStore {
	return &RedisStore{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
		ttl: ttl,
	}
}

func (s *RedisStore) key(entityType, entityID string) string {
	return fmt.Sprintf("rule:%s:%s", entityType, entityID)
}

func (s *RedisStore) Get(ctx context.Context, entityType, entityID string) (*Rule, bool) {
	key := s.key(entityType, entityID)

	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		// Normal cache miss
		observability.Debug(ctx, "redis cache miss",
			zap.String("key", key),
		)
		return nil, false
	}

	// Empty or null values should be treated as cache miss
	if val == "" || val == "null" || val == "{}" {
		observability.Warn(ctx, "redis returned empty rule",
			zap.String("key", key),
			zap.String("raw", val),
		)
		return nil, false
	}

	var r Rule
	if err := json.Unmarshal([]byte(val), &r); err != nil {
		observability.Warn(ctx, "redis cache unmarshal failed",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil, false
	}

	// Detect zero-value rule (invalid)
	if r.ID == "" && r.RuleType == "" && !r.ComplianceFlag {
		observability.Warn(ctx, "redis returned zero-value rule (treating as miss)",
			zap.String("key", key),
			zap.Any("rule", r),
		)
		return nil, false
	}

	observability.Debug(ctx, "redis cache hit",
		zap.String("key", key),
		zap.String("rule_id", r.ID),
		zap.String("rule_type", r.RuleType),
	)

	return &r, true
}

func (s *RedisStore) Set(ctx context.Context, entityType, entityID string, r *Rule) {
	key := s.key(entityType, entityID)

	b, err := json.Marshal(r)
	if err != nil {
		observability.Warn(ctx, "redis cache marshal failed",
			zap.String("key", key),
			zap.Error(err),
		)
		return
	}

	if err := s.client.Set(ctx, key, b, s.ttl).Err(); err != nil {
		observability.Warn(ctx, "redis cache set failed",
			zap.String("key", key),
			zap.Error(err),
		)
		return
	}

	observability.Debug(ctx, "redis cache set",
		zap.String("key", key),
		zap.String("rule_id", r.ID),
	)
}

func (s *RedisStore) Ping(ctx context.Context) error {
	return s.client.Ping(context.Background()).Err()
}
