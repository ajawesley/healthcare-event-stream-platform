package server

type panicError struct {
	panic any
}

func (e panicError) OK() any {
	return e.panic
}
