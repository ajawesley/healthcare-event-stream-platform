package transformer

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// asString safely converts any value from NormalizedEvent.Fields into a string.
// It handles nil, string, numeric, boolean, fmt.Stringer, JSON types, slices,
// maps, and pointers. It NEVER panics and NEVER logs.
func asString(v any) string {
	if v == nil {
		return ""
	}

	switch x := v.(type) {

	// Already a string
	case string:
		return x

	// Raw bytes
	case []byte:
		return string(x)

	// fmt.Stringer (time.Time, custom types, etc.)
	case fmt.Stringer:
		return x.String()

	// Common primitives
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		bool:
		return fmt.Sprintf("%v", x)

	// JSON numbers (from json.Unmarshal into map[string]any)
	case json.Number:
		return x.String()

	// Slice of primitives → JSON array string
	case []any:
		b, _ := json.Marshal(x)
		return string(b)

	// Map → JSON object string
	case map[string]any:
		b, _ := json.Marshal(x)
		return string(b)

	// Pointer → dereference safely
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr && !rv.IsNil() {
			return asString(rv.Elem().Interface())
		}

		// Fallback: best‑effort stringification
		b, err := json.Marshal(v)
		if err == nil {
			return string(b)
		}

		return fmt.Sprintf("%v", v)
	}
}
