package compliance

import (
	"context"
)

type MockClient struct {
	LookupFn func(ctx context.Context, entityType, entityID string) (*Rule, error)
	ReadyFn  func(ctx context.Context) error
	LiveFn   func() error
}

func (m *MockClient) LookupRule(ctx context.Context, entityType, entityID string) (*Rule, error) {
	return m.LookupFn(ctx, entityType, entityID)
}

func (m *MockClient) Ready(ctx context.Context) error {
	return m.ReadyFn(ctx)
}

func (m *MockClient) Live() error {
	return m.LiveFn()
}

func (m *MockClient) Close() {}
