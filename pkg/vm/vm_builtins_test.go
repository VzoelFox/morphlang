package vm

import (
	"testing"
	"github.com/VzoelFox/morphlang/pkg/object"
)

func TestBuiltinErrorHandling(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `galat("tes error")`,
			expected: object.NewError("tes error", "", 0, 0),
		},
		{
			input: `adalah_galat(galat("err"))`,
			expected: true,
		},
		{
			input: `adalah_galat(1)`,
			expected: false,
		},
		{
			input: `pesan_galat(galat("pesan rahasia"))`,
			expected: "pesan rahasia",
		},
		{
			input: `pesan_galat("bukan error")`,
			expected: object.NewError("argument to `pesan_galat` must be ERROR, got STRING", "E003", 0, 0),
		},
	}

	runVmTests(t, tests)
}

func TestVMRuntimeErrors(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `10 / 0`,
			expected: object.NewError("integer divide by zero", "", 0, 0),
		},
		{
			input: `"a" - "b"`,
			expected: object.NewError("string only supports add", "", 0, 0),
		},
		{
			input: `-"a"`,
			expected: object.NewError("minus not supported for type tag 3", "", 0, 0),
		},
		{
			input: `1 > "a"`,
			expected: object.NewError("unsupported comparison", "", 0, 0),
		},
	}

	runVmTests(t, tests)
}
