package resilience

import (
	"context"
	"time"

	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

type Dependency string

type Executor func(ctx context.Context) error
type Fallback func(ctx context.Context, err error) error

func Do(ctx context.Context, dep Dependency, exec Executor) error {
	return doInternal(ctx, dep, exec, nil)
}

func DoWithFallback(ctx context.Context, dep Dependency, exec Executor, fb Fallback) error {
	return doInternal(ctx, dep, exec, fb)
}

func doInternal(ctx context.Context, dep Dependency, exec Executor, fb Fallback) error {
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if !allow(dep) {
		observability.Warn(ctx, "circuit breaker open",
			zap.String("dependency", string(dep)),
		)
		if fb != nil {
			return fb(ctx, ErrCircuitOpen)
		}
		return ErrCircuitOpen
	}

	var err error

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			observability.Warn(ctx, "retrying dependency call",
				zap.String("dependency", string(dep)),
				zap.Int("attempt", attempt),
			)
			sleepWithBackoff(attempt)
		}

		if !acquireBulkhead(dep) {
			observability.Warn(ctx, "bulkhead full",
				zap.String("dependency", string(dep)),
			)
			if fb != nil {
				return fb(ctx, ErrBulkheadFull)
			}
			return ErrBulkheadFull
		}

		err = exec(ctx)
		releaseBulkhead(dep)

		if err == nil {
			recordSuccess(dep)

			// FIXED: correct call signature
			observability.ObserveDependencyLatency(
				ctx,
				string(dep), // depType
				"execute",   // operation
				string(dep), // target
				start,
			)

			observability.Info(ctx, "dependency call succeeded",
				zap.String("dependency", string(dep)),
				zap.Duration("latency_ms", time.Since(start)),
			)

			return nil
		}

		recordFailure(dep)

		observability.Warn(ctx, "dependency call failed",
			zap.String("dependency", string(dep)),
			zap.Error(err),
			zap.Int("attempt", attempt),
		)

		if !isRetryable(err) {
			break
		}
	}

	// FIXED: correct Error() signature
	observability.Error(ctx,
		"dependency call exhausted retries",
		err,
		"DEPENDENCY_CALL_EXHAUSTION_FAILURE",
		"external dependency repeatedly returned an error",
		zap.String("dependency", string(dep)),
	)

	if fb != nil {
		return fb(ctx, err)
	}

	return err
}
