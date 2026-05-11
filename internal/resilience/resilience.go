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

	// Hard timeout wrapper for the entire dependency call
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	// Circuit breaker check
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

		// Retry logging
		if attempt > 0 {
			observability.Warn(ctx, "retrying dependency call",
				zap.String("dependency", string(dep)),
				zap.Int("attempt", attempt),
			)
			sleepWithBackoff(attempt)
		}

		// Bulkhead guard
		if !acquireBulkhead(dep) {
			observability.Warn(ctx, "bulkhead full",
				zap.String("dependency", string(dep)),
			)
			if fb != nil {
				return fb(ctx, ErrBulkheadFull)
			}
			return ErrBulkheadFull
		}

		// Execute dependency in a goroutine to enforce timeout
		resultCh := make(chan error, 1)

		go func() {
			resultCh <- exec(ctx)
		}()

		select {
		case err = <-resultCh:
			// Dependency returned
			releaseBulkhead(dep)

			if err == nil {
				recordSuccess(dep)

				// Correct observability signature
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

			// Dependency returned an error
			recordFailure(dep)

			observability.Warn(ctx, "dependency call failed",
				zap.String("dependency", string(dep)),
				zap.Error(err),
				zap.Int("attempt", attempt),
			)

			if !isRetryable(err) {
				// Non-retryable error → break immediately
				attempt = 3
			}

		case <-ctx.Done():
			// Timeout or cancellation
			releaseBulkhead(dep)
			recordFailure(dep)

			err = ctx.Err()

			observability.Warn(ctx, "dependency call timed out",
				zap.String("dependency", string(dep)),
				zap.Error(err),
				zap.Int("attempt", attempt),
			)

			// Timeout is retryable for first 2 attempts
			if attempt == 2 {
				break
			}
		}
	}

	// Exhausted retries
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
