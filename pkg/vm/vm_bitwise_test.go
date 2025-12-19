package vm

import (
	"testing"

	"github.com/VzoelFox/morphlang/pkg/object"
)

func TestBitwiseIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1 & 1", 1},
		{"1 & 0", 0},
		{"0 & 1", 0},
		{"0 & 0", 0},
		{"5 & 3", 1},
		{"7 & 3", 3},

		{"1 | 1", 1},
		{"1 | 0", 1},
		{"0 | 1", 1},
		{"0 | 0", 0},
		{"5 | 3", 7},
		{"4 | 2", 6},

		{"1 ^ 1", 0},
		{"1 ^ 0", 1},
		{"0 ^ 1", 1},
		{"0 ^ 0", 0},
		{"5 ^ 3", 6},

		{"~1", -2},
		{"~0", -1},
		{"~(-2)", 1},

		{"1 << 1", 2},
		{"1 << 2", 4},
		{"5 << 1", 10},
		{"10 << 2", 40},

		{"2 >> 1", 1},
		{"4 >> 1", 2},
		{"10 >> 1", 5},
		{"40 >> 2", 10},
		{"-4 >> 1", -2},

		{"5 + 2 * 3", 11},
		{"5 & 2 | 1", 1},
		{"5 | 2 & 1", 5},
		{"1 << 2 + 1", 8},
		{"10 * 2 >> 1", 10},
	}

	runVmTests(t, tests)
}

func TestBitwiseErrors(t *testing.T) {
	tests := []vmTestCase{
		{
			"benar & salah",
			object.NewError("unsupported types for bitwise operation: type tag 2 type tag 2", "", 0, 0),
		},
		{
			"1 | \"2\"",
			object.NewError("unsupported types for bitwise operation: type tag 1 type tag 3", "", 0, 0),
		},
		{
			"~benar",
			object.NewError("bitnot not supported for type tag 2", "", 0, 0),
		},
		{
			"1 << -1",
			object.NewError("negative shift count", "", 0, 0),
		},
		{
			"1 >> -1",
			object.NewError("negative shift count", "", 0, 0),
		},
	}

	runVmTests(t, tests)
}
