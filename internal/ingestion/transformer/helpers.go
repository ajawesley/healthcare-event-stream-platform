package transformer

import "fmt"

// asString safely converts any value from NormalizedEvent.Fields into a string.
// It handles nil, string, numeric, boolean, and fmt.Stringer types.
func asString(v any) string {
	if v == nil {
		return ""
	}

	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	case fmt.Stringer:
		return x.String()
	case int, int32, int64, float32, float64, bool:
		return fmt.Sprintf("%v", x)
	default:
		return ""
	}
}
