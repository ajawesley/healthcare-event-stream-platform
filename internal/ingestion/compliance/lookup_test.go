package compliance

import (
	"context"
	"errors"
	"testing"
)

func TestLookupRule(t *testing.T) {
	tests := []struct {
		name     string
		mockRule *Rule
		mockErr  error
		wantErr  error
	}{
		{
			name: "success",
			mockRule: &Rule{
				ID:             "rule-1",
				EntityType:     "patient",
				EntityID:       "123",
				RuleType:       "REGULATORY",
				ComplianceFlag: true,
				ReasonCode:     "OK",
			},
			mockErr: nil,
			wantErr: nil,
		},
		{
			name:     "not found",
			mockRule: nil,
			mockErr:  ErrNotFound,
			wantErr:  ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mock := &MockClient{
				LookupFn: func(ctx context.Context, entityType, entityID string) (*Rule, error) {
					return tt.mockRule, tt.mockErr
				},
			}

			rule, err := mock.LookupRule(context.Background(), "patient", "123")

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if tt.mockRule != nil && rule.ID != tt.mockRule.ID {
				t.Fatalf("expected rule ID %s, got %s", tt.mockRule.ID, rule.ID)
			}
		})
	}
}
