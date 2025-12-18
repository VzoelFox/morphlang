package vm

import (
	"testing"
)

func TestFloatArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"1.0", 1.0},
		{"2.5", 2.5},
		{"1.0 + 2.0", 3.0},
		{"1.0 - 2.0", -1.0},
		{"1.5 * 2.0", 3.0},
		{"4.0 / 2.0", 2.0},
		{"-5.5", -5.5},
		// Mixed Types
		{"1 + 2.0", 3.0},
		{"2.5 + 1", 3.5},
		{"2.0 * 2", 4.0},
		{"10 / 2.0", 5.0},
		// Comparisons
		{"1.0 < 2.0", true},
		{"1.0 > 2.0", false},
		{"1.0 == 1.0", true},
		{"1.0 != 2.0", true},
		{"1.0 == 1", true}, // Mixed Comparison
		{"1 == 1.0", true},
		{"1 < 2.0", true},
		{"2.5 > 2", true},
	}

	runVmTests(t, tests)
}

func TestStringConcatenationMixed(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"a" + 1`, "a1"},
		{`"a" + 1.5`, "a1.5"},
		{`1 + "a"`, "1a"},
		{`1.5 + "a"`, "1.5a"},
		{`" " + 1`, " 1"},
	}

	runVmTests(t, tests)
}
