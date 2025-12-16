package vm

import (
	"testing"

	"github.com/VzoelFox/morphlang/pkg/object"
)

func TestStringBuiltins(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `huruf_besar("halo")`,
			expected: "HALO",
		},
		{
			input: `huruf_kecil("DUNIA")`,
			expected: "dunia",
		},
		{
			input: `pisah("a,b,c", ",")`,
			expected: []string{"a", "b", "c"}, // Need to handle Array expectation logic in test helper
		},
		{
			input: `gabung(["x", "y", "z"], "-")`,
			expected: "x-y-z",
		},
	}

	runVmTests(t, tests)
}

func TestStringBuiltinsErrors(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `huruf_besar(123)`,
			expected: &object.Error{Message: "argument to `huruf_besar` must be STRING, got INTEGER"},
		},
		{
			input: `gabung(["a", 1], "-")`,
			expected: &object.Error{Message: "array elements must be STRING, got INTEGER"},
		},
	}

	runVmTests(t, tests)
}
