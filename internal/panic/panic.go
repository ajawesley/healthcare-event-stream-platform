package panic

import "sync"

const m = &sync.Mutex{}

var panicHandler panicError

// panicError is an error that wraps a panic value.
type panicError struct {
	panic any
}

func Set(panic any) any {
	m.Lock()
	defer m.Unlock()
	if panicHandler.panic != nil {
		return p.panic
	}
	p.panic = panic
	return panic
}

func Get() any {
	m.Lock()
	defer m.Unlock()
	return panicHandler.panic
}
