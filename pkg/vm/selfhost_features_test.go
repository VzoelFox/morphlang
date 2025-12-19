package vm

import (
	"testing"
)

func TestSelfHostFeatures(t *testing.T) {
	tests := []vmTestCase{
		// 1. String Indexing
		{
			input: `s = "abc"; s[0]`,
			expected: "a",
		},
		{
			input: `s = "Halo"; s[3]`,
			expected: "o",
		},
		{
			input: `s = "Out"; s[100]`, // Out of bounds -> Null
			expected: Null,
		},

		// 2. chr() & ord()
		{
			input: `chr(65)`,
			expected: "A",
		},
		{
			input: `ord("A")`,
			expected: 65,
		},
		{
			input: `chr(ord("X"))`,
			expected: "X",
		},

		// 3. Byte Manipulation
		// Use `alokasi` to get buffer
		{
			input: `
				p = alokasi(10)
				tulis_byte(p, 0, 65) # 'A'
				tulis_byte(p, 1, 66) # 'B'
				baca_byte(p, 1)
			`,
			expected: 66,
		},
		{
			input: `
				p = alokasi(10)
				tulis_byte(p, 0, 72) # 'H'
				tulis_byte(p, 1, 105) # 'i'
				baca_mem(p, 2)
			`,
			expected: "Hi",
		},
	}

	runVmTests(t, tests)
}
