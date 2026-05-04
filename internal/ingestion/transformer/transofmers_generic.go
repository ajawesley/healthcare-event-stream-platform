package transformer

// genericTransformer is a no‑op transformer used for FormatGeneric.
type genericTransformer struct{}

func NewGenericTransformer() Transformer {
	return &genericTransformer{}
}

func (t *genericTransformer) Transform(value any) (any, error) {
	return value, nil
}
