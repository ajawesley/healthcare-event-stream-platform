package transformer

// Transformer defines the interface for format‑specific transformations.
// Each transformer receives the parsed payload (already routed by Router)
// and returns a canonical representation or an error.
type Transformer interface {
	Transform(value any) (any, error)
}
