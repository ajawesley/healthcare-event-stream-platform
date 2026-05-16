package observability

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

// -----------------------------------------------------------------------------
// LineageStage
// -----------------------------------------------------------------------------

type LineageStage struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

// -----------------------------------------------------------------------------
// Lineage
// -----------------------------------------------------------------------------

type Lineage struct {
	TraceID         string    `json:"trace_id"`
	EventID         string    `json:"event_id"`
	IngestTimestamp time.Time `json:"ingest_timestamp"`

	stages []LineageStage

	ticker *time.Ticker
	done   chan struct{}

	mu sync.RWMutex
}

// -----------------------------------------------------------------------------
// Global Deadlock State
// -----------------------------------------------------------------------------

// Once ANY lineage exceeds the timeout, the entire service is considered
// deadlocked until restart. This matches the liveness/readiness contract.
var (
	deadlocked bool
	dmu        sync.RWMutex
)

func DeadLocked() bool {
	dmu.RLock()
	defer dmu.RUnlock()
	return deadlocked
}

// -----------------------------------------------------------------------------
// internal: startDeadlockTimer
// -----------------------------------------------------------------------------

func (l *Lineage) startDeadlockTimer() {
	if l == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.ticker != nil {
		return
	}

	l.ticker = time.NewTicker(1 * time.Minute)
	l.done = make(chan struct{})

	t := l.ticker
	doneCh := l.done

	go func() {
		for {
			select {
			case <-t.C:
				dmu.Lock()
				deadlocked = true
				dmu.Unlock()
				return

			case <-doneCh:
				return
			}
		}
	}()
}

// -----------------------------------------------------------------------------
// internal: resetDeadlockTimer
// -----------------------------------------------------------------------------

func (l *Lineage) resetDeadlockTimer() {
	if l == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.ticker == nil {
		return
	}

	// Drain pending ticks to avoid false deadlocks
	for {
		select {
		case <-l.ticker.C:
		default:
			goto drained
		}
	}
drained:

	l.ticker.Reset(1 * time.Minute)
}

// -----------------------------------------------------------------------------
// Complete — stops timer and cleans up resources
// -----------------------------------------------------------------------------

// Complete is safe on a nil receiver.
func (l *Lineage) Complete() {
	if l == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.ticker != nil {
		l.ticker.Stop()
		l.ticker = nil
	}

	if l.done != nil {
		close(l.done)
		l.done = nil
	}

	// IMPORTANT:
	// We DO NOT clear the global deadlocked flag here.
	// Once a deadlock is detected, the service is considered deadlocked
	// until the container is restarted.
}

// -----------------------------------------------------------------------------
// Stage API
// -----------------------------------------------------------------------------

// MarkStage is safe on a nil receiver.
func (l *Lineage) MarkStage(name string) {
	if l == nil {
		return
	}

	now := time.Now().UTC()

	l.mu.Lock()
	l.stages = append(l.stages, LineageStage{
		Name:      name,
		Timestamp: now,
	})
	count := len(l.stages)
	l.mu.Unlock()

	if count == 1 {
		l.startDeadlockTimer()
	} else {
		l.resetDeadlockTimer()
	}
}

func (l *Lineage) Stages() []LineageStage {
	if l == nil {
		return nil
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	out := make([]LineageStage, len(l.stages))
	copy(out, l.stages)
	return out
}

func (l *Lineage) HasStage(name string) bool {
	if l == nil {
		return false
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, s := range l.stages {
		if s.Name == name {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// Context Key
// -----------------------------------------------------------------------------

type lineageKeyType struct{}

var lineageKey = lineageKeyType{}

// -----------------------------------------------------------------------------
// NewLineage(ctx)
// -----------------------------------------------------------------------------

func NewLineage(ctx context.Context) (*Lineage, context.Context) {
	if ctx == nil {
		return &Lineage{
			EventID:         uuid.NewString(),
			IngestTimestamp: time.Now().UTC(),
			stages:          make([]LineageStage, 0, 100),
		}, nil
	}

	if existing := GetLineage(ctx); existing != nil {
		return existing, ctx
	}

	var traceID string
	if span := trace.SpanFromContext(ctx); span != nil {
		sc := span.SpanContext()
		if sc.IsValid() {
			traceID = sc.TraceID().String()
		}
	}

	l := &Lineage{
		TraceID:         traceID,
		EventID:         uuid.NewString(),
		IngestTimestamp: time.Now().UTC(),
		stages:          make([]LineageStage, 0, 100),
	}

	ctx = context.WithValue(ctx, lineageKey, l)
	return l, ctx
}

// -----------------------------------------------------------------------------
// InjectLineage
// -----------------------------------------------------------------------------

func InjectLineage(ctx context.Context, l *Lineage) context.Context {
	if ctx == nil || l == nil {
		return ctx
	}
	return context.WithValue(ctx, lineageKey, l)
}

// -----------------------------------------------------------------------------
// GetLineage
// -----------------------------------------------------------------------------

func GetLineage(ctx context.Context) *Lineage {
	if ctx == nil {
		return nil
	}
	val := ctx.Value(lineageKey)
	if val == nil {
		return nil
	}
	return val.(*Lineage)
}
