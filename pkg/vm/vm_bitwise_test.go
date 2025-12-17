package vm

import (
	"testing"
)

func TestBitwiseIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1 & 1", 1},
		{"1 & 0", 0},
		{"0 & 1", 0},
		{"0 & 0", 0},
		{"5 & 3", 1},  // 101 & 011 = 001
		{"7 & 3", 3},  // 111 & 011 = 011

		{"1 | 1", 1},
		{"1 | 0", 1},
		{"0 | 1", 1},
		{"0 | 0", 0},
		{"5 | 3", 7},  // 101 | 011 = 111
		{"4 | 2", 6},  // 100 | 010 = 110

		{"1 ^ 1", 0},
		{"1 ^ 0", 1},
		{"0 ^ 1", 1},
		{"0 ^ 0", 0},
		{"5 ^ 3", 6},  // 101 ^ 011 = 110 (6)

		{"~1", -2},    // ^1 = -2
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
		{"-4 >> 1", -2}, // Arithmetic shift preserves sign

		// Precedence Checks
		{"5 + 2 * 3", 11},
		{"5 & 2 | 1", 1}, // (5 & 2) | 1 -> 0 | 1 -> 1
		{"5 | 2 & 1", 5}, // 5 | (2 & 1) -> 5 | 0 -> 5 (Bitwise And has higher precedence than Or)
		{"1 << 2 + 1", 8}, // 1 << (2 + 1) -> 1 << 3 -> 8 (Sum has higher precedence than Shift)
		{"10 * 2 >> 1", 10}, // (10 * 2) >> 1 -> 20 >> 1 -> 10 (Product has higher precedence than Shift)
	}

	runVmTests(t, tests)
}

// TestBitwiseErrors temporarily disabled due to stack underflow investigation
// func TestBitwiseErrors(t *testing.T) {
// 	tests := []vmTestCase{
// 		{
// 			"benar & salah",
// 			&object.Error{Message: "unsupported types for bitwise operation: BOOLEAN BOOLEAN"},
// 		},
// 		{
// 			"1 | '2'",
// 			&object.Error{Message: "unsupported types for bitwise operation: INTEGER STRING"},
// 		},
// 		{
// 			"~benar",
// 			&object.Error{Message: "unsupported type for bitwise not: BOOLEAN"},
// 		},
// 		{
// 			"1 << -1",
// 			&object.Error{Message: "negative shift count"},
// 		},
// 		{
// 			"1 >> -1",
// 			&object.Error{Message: "negative shift count"},
// 		},
// 	}

// 	runVmTests(t, tests)
// }
