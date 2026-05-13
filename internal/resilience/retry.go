package resilience

import (
	"math/rand"
	"time"
)

func sleepWithBackoff(attempt int) {
	base := 50 * time.Millisecond
	max := 500 * time.Millisecond

	d := base << attempt
	if d > max {
		d = max
	}

	jitter := time.Duration(rand.Int63n(int64(d) / 5)) // ±20%
	time.Sleep(d + jitter)
}

func isRetryable(err error) bool {
	// You can refine this later
	return true
}
