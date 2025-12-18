package vm

import (
	"testing"
	"github.com/VzoelFox/morphlang/pkg/object"
)

func TestBuiltinErrorHandling(t *testing.T) {
	tests := []vmTestCase{
		// Test 'galat'
		{
			input: `galat("tes error")`,
			expected: &object.Error{Message: "tes error"},
		},
		// Test 'adalah_galat'
		{
			input: `adalah_galat(galat("err"))`,
			expected: true,
		},
		{
			input: `adalah_galat(1)`,
			expected: false,
		},
		// Test 'pesan_galat'
		{
			input: `pesan_galat(galat("pesan rahasia"))`,
			expected: "pesan rahasia",
		},
		{
			input: `pesan_galat("bukan error")`,
			expected: &object.Error{Message: "argument to `pesan_galat` must be ERROR, got STRING"},
		},
	}

	runVmTests(t, tests)
}

func TestVMRuntimeErrors(t *testing.T) {
	tests := []vmTestCase{
		// Division by zero
		{
			input: `10 / 0`,
			expected: &object.Error{Message: "division by zero"},
		},
		// Type mismatch (Arithmetic)
		// 1 + "a" is now valid ("1a")
		{
			input: `"a" - "b"`,
			expected: &object.Error{Message: "unknown string operator: 33"},
		},
		{
			input: `-"a"`,
			expected: &object.Error{Message: "unsupported type for negation: STRING"},
		},
		// Type mismatch (Comparison)
		{
			input: `1 > "a"`,
			expected: &object.Error{Message: "unsupported comparison: INTEGER 38 STRING"},
		},
	}

	runVmTests(t, tests)
}

func testExpectedObjectIncludesError(t *testing.T, obj object.Object, expected interface{}) {
	// Custom helper if needed, but runVmTests calls testExpectedObject which handles exact match.
	// For errors, testExpectedObject checks pointer equality? No, it casts.
	// Let's rely on testExpectedObject extending.
}
