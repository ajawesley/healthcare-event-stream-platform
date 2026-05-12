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
type RetryPolicy func(err error) bool

func Do(ctx context.Context, dep Dependency, exec Executor, retry RetryPolicy) error {
	return doInternal(ctx, dep, exec, nil, retry)
}

func DoWithFallback(ctx context.Context, dep Dependency, exec Executor, fb Fallback, retry RetryPolicy) error {
	return doInternal(ctx, dep, exec, fb, retry)
}

func doInternal(ctx context.Context, dep Dependency, exec Executor, fb Fallback, retry RetryPolicy) error {
	start := time.Now()

	// Hard timeout for entire dependency call
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	// Circuit breaker
	if !allow(dep) {
		observability.Warn(ctx, "circuit breaker open",
			zap.String("dependency", string(dep)),
		)
		if fb != nil {
			return fb(ctx, ErrCircuitOpen)
		}
		return ErrCircuitOpen
	}

	// Default retry policy = retry everything
	if retry == nil {
		retry = func(err error) bool { return true }
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

		// Bulkhead
		if !acquireBulkhead(dep) {
			observability.Warn(ctx, "bulkhead full",
				zap.String("dependency", string(dep)),
			)
			if fb != nil {
				return fb(ctx, ErrBulkheadFull)
			}
			return ErrBulkheadFull
		}

		// Execute dependency in goroutine
		resultCh := make(chan error, 1)
		go func() {
			resultCh <- exec(ctx)
		}()

		select {
		case err = <-resultCh:
			releaseBulkhead(dep)

			if err == nil {
				recordSuccess(dep)

				observability.ObserveDependencyLatency(
					ctx,
					string(dep),
					"execute",
					string(dep),
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

			// Caller decides retryability (default = true)
			if !retry(err) {
				attempt = 3
			}

		case <-ctx.Done():
			releaseBulkhead(dep)
			recordFailure(dep)

			err = ctx.Err()

			observability.Warn(ctx, "dependency call timed out",
				zap.String("dependency", string(dep)),
				zap.Error(err),
				zap.Int("attempt", attempt),
			)

			// Timeout retryability also controlled by caller (default = true)
			if !retry(err) {
				attempt = 3
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
