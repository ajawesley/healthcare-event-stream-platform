package compliance

import (
	"context"

	"github.com/ajawes/hesp/internal/observability"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.uber.org/zap"
)

type DynamoStore struct {
	client *dynamodb.Client
	table  string
}

func NewDynamoStore(client *dynamodb.Client, table string) *DynamoStore {
	return &DynamoStore{
		client: client,
		table:  table,
	}
}

func (s *DynamoStore) Lookup(ctx context.Context, entityType, entityID string) (*Rule, error) {
	observability.Debug(ctx, "dynamodb compliance lookup invoked",
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID),
	)

	out, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &s.table,
		Key: map[string]types.AttributeValue{
			"entity_type": &types.AttributeValueMemberS{Value: entityType},
			"entity_id":   &types.AttributeValueMemberS{Value: entityID},
		},
	})
	if err != nil {
		observability.Warn(ctx, "dynamodb get item failed",
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityID),
			zap.Error(err),
		)
		return nil, err
	}

	if out.Item == nil {
		observability.Warn(ctx, "no compliance rule found in dynamodb",
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityID),
		)
		return nil, ErrNotFound
	}

	var r Rule
	if err := attributevalue.UnmarshalMap(out.Item, &r); err != nil {
		observability.Warn(ctx, "dynamodb unmarshal failed",
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityID),
			zap.Error(err),
		)
		return nil, err
	}

	observability.Info(ctx, "dynamodb compliance rule retrieved",
		zap.String("rule_id", r.ID),
		zap.String("rule_type", r.RuleType),
		zap.Bool("flag", r.ComplianceFlag),
		zap.String("reason", r.ReasonCode),
	)

	return &r, nil
}
