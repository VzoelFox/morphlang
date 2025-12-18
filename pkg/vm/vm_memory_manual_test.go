package vm

import (
	"testing"
)

func TestManualMemory(t *testing.T) {
	// 1. Allocate and Write/Read
	input := `
    p = alokasi(10);
    tulis_mem(p, "Morph");
    d = baca_mem(p, 5);
    d
    `
	runVmTest(t, input, "Morph")

	// 2. Pointer Arithmetic (Simulated via casts)
	inputPtrMath := `
    p = alokasi(10);
    tulis_mem(p, "Morph");

    addr = alamat(p);
    addr2 = addr + 1;
    p2 = ptr_dari(addr2);

    d = baca_mem(p2, 4);
    d
    `
	// "orph"
	runVmTest(t, inputPtrMath, "orph")
}
