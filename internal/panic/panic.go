package panic

import "sync"

var (
	m            = &sync.Mutex{}
	panicHandler panicError
)

// panicError is an error that wraps a panic value.
type panicError struct {
	panic any
}

func Set(panic any) any {
	m.Lock()
	defer m.Unlock()
	if panicHandler.panic != nil {
		return panicHandler.panic
	}
	panicHandler.panic = panic
	return panic
}

func Get() any {
	m.Lock()
	defer m.Unlock()
	return panicHandler.panic
}
