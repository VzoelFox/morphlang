package vm

import (
	"testing"
)

func TestConcurrencyScaffolding(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			ch = saluran_baru(0)
			luncurkan(fungsi()
				kirim(ch, 42)
			akhir)
			terima(ch)
			`,
			expected: 42,
		},
		{
			input: `
			ch = saluran_baru(2)
			luncurkan(fungsi()
				kirim(ch, 10)
				kirim(ch, 20)
			akhir)
			a = terima(ch)
			b = terima(ch)
			a + b
			`,
			expected: 30,
		},
	}

	runVmTests(t, tests)
}

func TestConcurrencyClosureCapture(t *testing.T) {
	// Test if spawned function captures variables (via Closure)
	tests := []vmTestCase{
		{
			input: `
			x = 100
			ch = saluran_baru(0)
			luncurkan(fungsi()
				kirim(ch, x)
			akhir)
			terima(ch)
			`,
			expected: 100,
		},
	}

	runVmTests(t, tests)
}
