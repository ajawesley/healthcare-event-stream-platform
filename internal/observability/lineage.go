package observability

import (
	"context"
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
// Lineage (stages unexported)
// -----------------------------------------------------------------------------

type Lineage struct {
	TraceID         string    `json:"trace_id"`
	EventID         string    `json:"event_id"`
	IngestTimestamp time.Time `json:"ingest_timestamp"`

	// unexported to prevent external mutation
	stages []LineageStage
}

// -----------------------------------------------------------------------------
// Stage API
// -----------------------------------------------------------------------------

func (l *Lineage) MarkStage(name string) {
	l.stages = append(l.stages, LineageStage{
		Name:      name,
		Timestamp: time.Now().UTC(),
	})
}

func (l *Lineage) Stages() []LineageStage {
	out := make([]LineageStage, len(l.stages))
	copy(out, l.stages)
	return out
}

func (l *Lineage) HasStage(name string) bool {
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
	// ctx nil → lineage exists, context optional
	if ctx == nil {
		return &Lineage{
			EventID:         uuid.NewString(),
			IngestTimestamp: time.Now().UTC(),
		}, nil
	}

	// idempotent
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
