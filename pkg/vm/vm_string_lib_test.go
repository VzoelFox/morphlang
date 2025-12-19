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
			expected: []string{"a", "b", "c"},
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
			expected: object.NewError("argument to `huruf_besar` must be STRING, got INTEGER", "", 0, 0),
		},
		{
			input: `gabung(["a", 1], "-")`,
			expected: object.NewError("array elements must be STRING, got INTEGER", "", 0, 0),
		},
	}

	runVmTests(t, tests)
}
