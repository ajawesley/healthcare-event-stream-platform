package compliance

import (
	"context"
	"testing"
	"time"

	"github.com/ajawes/hesp/internal/ingestion/models"
)

func TestGuardApply(t *testing.T) {
	tests := []struct {
		name       string
		mockRule   *Rule
		mockErr    error
		wantFlag   bool
		wantReason string
	}{
		{
			name: "rule found",
			mockRule: &Rule{
				ID:             "r1",
				RuleType:       "REG",
				ComplianceFlag: true,
				ReasonCode:     "OK",
			},
			mockErr:    nil,
			wantFlag:   true,
			wantReason: "OK",
		},
		{
			name:       "no rule",
			mockRule:   nil,
			mockErr:    ErrNotFound,
			wantFlag:   false,
			wantReason: "NO_RULE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mock := &MockClient{
				LookupFn: func(ctx context.Context, entityType, entityID string) (*Rule, error) {
					return tt.mockRule, tt.mockErr
				},
			}

			guard := NewGuard(mock, NewCircuitBreaker(5, time.Second))

			evt := &models.CanonicalEvent{
				Patient: &models.CanonicalPatient{ID: "123"},
			}

			err := guard.Apply(context.Background(), evt)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if evt.ComplianceFlag != tt.wantFlag {
				t.Fatalf("expected flag %v, got %v", tt.wantFlag, evt.ComplianceFlag)
			}

			if evt.ComplianceReason != tt.wantReason {
				t.Fatalf("expected reason %s, got %s", tt.wantReason, evt.ComplianceReason)
			}
		})
	}
}
