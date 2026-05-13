package resilience

import (
	"context"
	"sync"
	"time"

	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

type state int

const (
	stateClosed state = iota
	stateOpen
	stateHalfOpen
)

type breaker struct {
	mu       sync.Mutex
	failures int
	state    state
	openedAt time.Time
}

var (
	breakers   = make(map[Dependency]*breaker)
	breakersMu sync.Mutex
)

func getBreaker(dep Dependency) *breaker {
	breakersMu.Lock()
	defer breakersMu.Unlock()
	if b, ok := breakers[dep]; ok {
		return b
	}
	b := &breaker{}
	breakers[dep] = b
	return b
}

func allow(dep Dependency) bool {
	b := getBreaker(dep)
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case stateClosed:
		return true

	case stateOpen:
		if time.Since(b.openedAt) > 30*time.Second {
			b.state = stateHalfOpen
			return true
		}
		return false

	case stateHalfOpen:
		return true
	}

	return true
}

func recordSuccess(dep Dependency) {
	b := getBreaker(dep)
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state != stateClosed {
		observability.Info(context.Background(), "circuit breaker closed",
			zap.String("dependency", string(dep)),
		)
	}

	b.failures = 0
	b.state = stateClosed
}

func recordFailure(dep Dependency) {
	b := getBreaker(dep)
	b.mu.Lock()
	defer b.mu.Unlock()

	// If already open, nothing to do
	if b.state == stateOpen {
		return
	}

	// If half-open → immediately open again
	if b.state == stateHalfOpen {
		b.state = stateOpen
		b.openedAt = time.Now()

		observability.Warn(context.Background(), "circuit breaker re-opened from half-open",
			zap.String("dependency", string(dep)),
		)
		return
	}

	// Closed state → increment failures
	b.failures++

	if b.failures >= 3 {
		b.state = stateOpen
		b.openedAt = time.Now()

		observability.Warn(context.Background(), "circuit breaker opened",
			zap.String("dependency", string(dep)),
			zap.Int("failures", b.failures),
		)
	}
}
