package vm

import (
	"testing"
)

func TestClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			fungsi luar(a)
				fungsi dalam()
					kembalikan a + 10
				akhir
				kembalikan dalam()
			akhir
			luar(5)
			`,
			expected: 15,
		},
		{
			input: `
			fungsi luar(a)
				fungsi dalam(b)
					kembalikan a + b
				akhir
				kembalikan dalam(7)
			akhir
			luar(3)
			`,
			expected: 10,
		},
		{
			input: `
			fungsi luar(a)
				fungsi dalam(b)
					fungsi lebih_dalam(c)
						kembalikan a + b + c
					akhir
					kembalikan lebih_dalam
				akhir
				kembalikan dalam
			akhir
			luar(1)(2)(3)
			`,
			expected: 6,
		},
		{
			input: `
			fungsi buat_penutup()
				x = 50
				fungsi penutup()
					kembalikan x
				akhir
				kembalikan penutup
			akhir
			penutup = buat_penutup()
			penutup()
			`,
			expected: 50,
		},
	}

	runVmTests(t, tests)
}

func TestRecursiveClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			fungsi hitung_mundur(x)
				jika (x == 0)
					kembalikan 0
				lainnya
					kembalikan hitung_mundur(x - 1)
				akhir
			akhir
			hitung_mundur(1)
			`,
			expected: 0,
		},
		{
			input: `
			fungsi buat_penutup(x)
				fungsi penutup()
					kembalikan x
				akhir
				kembalikan penutup
			akhir
			penutup = buat_penutup(99)
			penutup()
			`,
			expected: 99,
		},
	}

	runVmTests(t, tests)
}
