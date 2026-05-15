package compliance

import (
	"context"
	"fmt"
)

type mockClient struct{}

func NewMockClient() ClientAPI {
	return &mockClient{}
}

func (m *mockClient) LookupRule(ctx context.Context, entityType, entityID string) (*Rule, error) {
	// Always return "no rule found" so fallback logic runs
	return nil, fmt.Errorf("mock: no rule found for %s/%s", entityType, entityID)
}
