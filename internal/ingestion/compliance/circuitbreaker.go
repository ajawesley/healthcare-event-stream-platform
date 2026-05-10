package compliance

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	mu        sync.Mutex
	failures  int
	threshold int
	openUntil time.Time
	cooldown  time.Duration
}

func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		cooldown:  cooldown,
	}
}

func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if time.Now().Before(cb.openUntil) {
		observability.Warn(context.Background(), "circuit breaker open",
			zap.Time("open_until", cb.openUntil),
		)
		return ErrCircuitOpen
	}

	return nil
}

func (cb *CircuitBreaker) Success() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	observability.Debug(context.Background(), "circuit breaker success reset")
}

func (cb *CircuitBreaker) Failure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	observability.Warn(context.Background(), "circuit breaker failure increment",
		zap.Int("failures", cb.failures),
	)

	if cb.failures >= cb.threshold {
		cb.openUntil = time.Now().Add(cb.cooldown)
		observability.Warn(context.Background(), "circuit breaker opened",
			zap.Time("open_until", cb.openUntil),
		)
	}
}
