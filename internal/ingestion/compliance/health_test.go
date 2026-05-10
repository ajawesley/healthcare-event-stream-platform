package compliance

import (
	"context"
	"errors"
	"testing"
)

func TestClientLive(t *testing.T) {
	mock := &MockClient{
		LiveFn: func() error { return nil },
	}

	if err := mock.Live(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestClientReady(t *testing.T) {
	mock := &MockClient{
		ReadyFn: func(ctx context.Context) error { return nil },
	}

	if err := mock.Ready(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	mockFail := &MockClient{
		ReadyFn: func(ctx context.Context) error { return errors.New("db down") },
	}

	if err := mockFail.Ready(context.Background()); err == nil {
		t.Fatalf("expected error, got nil")
	}
}
