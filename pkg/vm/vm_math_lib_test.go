package vm

import (
	"testing"

	"github.com/VzoelFox/morphlang/pkg/object"
)

func TestMathBuiltins(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `abs(-5)`,
			expected: 5,
		},
		{
			input: `max(10, 20)`,
			expected: 20,
		},
		{
			input: `min(10, 20)`,
			expected: 10,
		},
		{
			input: `pow(2, 3)`,
			expected: 8,
		},
		{
			input: `sqrt(16)`,
			expected: 4,
		},
	}

	runVmTests(t, tests)
}

func TestMathBuiltinsErrors(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `sqrt(-1)`,
			expected: &object.Error{Message: "cannot calculate square root of negative number"},
		},
		{
			input: `pow(2, "3")`,
			expected: &object.Error{Message: "second argument to `pow` must be INTEGER, got STRING"},
		},
	}

	runVmTests(t, tests)
}
