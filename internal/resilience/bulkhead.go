package resilience

import "sync"

type bulkhead struct {
	ch chan struct{}
}

var (
	bulkheads   = map[Dependency]*bulkhead{}
	bulkheadsMu = &sync.Mutex{}
)

func getBulkhead(dep Dependency) *bulkhead {
	bulkheadsMu.Lock()
	defer bulkheadsMu.Unlock()

	if b, ok := bulkheads[dep]; ok {
		return b
	}

	b := &bulkhead{ch: make(chan struct{}, 10)}
	bulkheads[dep] = b
	return b
}

func acquireBulkhead(dep Dependency) bool {
	b := getBulkhead(dep)
	select {
	case b.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

func releaseBulkhead(dep Dependency) {
	b := getBulkhead(dep)
	select {
	case <-b.ch:
	default:
	}
}
