package resilience

import "errors"

var (
	ErrCircuitOpen  = errors.New("circuit open")
	ErrBulkheadFull = errors.New("bulkhead full")
)
