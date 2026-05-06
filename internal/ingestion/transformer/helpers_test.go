package transformer

import (
	"encoding/json"
	"testing"
)

type stringerType struct {
	v string
}

func (s stringerType) String() string { return "stringer:" + s.v }

func TestAsString(t *testing.T) {
	num := 42
	strPtr := &num

	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"bytes", []byte("world"), "world"},
		{"int", 123, "123"},
		{"float", 3.14, "3.14"},
		{"bool_true", true, "true"},
		{"bool_false", false, "false"},
		{"json_number", json.Number("55.2"), "55.2"},
		{"stringer", stringerType{"abc"}, "stringer:abc"},
		{"pointer_to_int", strPtr, "42"},
		{"slice_of_any", []any{"a", 1, true}, `["a",1,true]`},
		{"map_string_any", map[string]any{"x": 1, "y": "z"}, `{"x":1,"y":"z"}`},
		{"unsupported_struct", struct{ A int }{A: 9}, `{"A":9}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := asString(tt.input)
			if out != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, out)
			}
		})
	}
}
