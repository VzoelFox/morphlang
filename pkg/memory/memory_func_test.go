package memory

import (
	"bytes"
	"testing"
)

func TestAllocCompiledFunction(t *testing.T) {
	InitCabinet()

	instr := []byte{0x01, 0x02, 0x03}
	ptr, err := AllocCompiledFunction(instr, 5, 2)
	if err != nil {
		t.Fatalf("Alloc failed: %v", err)
	}

	readInstr, locals, params, err := ReadCompiledFunction(ptr)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if locals != 5 {
		t.Errorf("Locals: want 5, got %d", locals)
	}
	if params != 2 {
		t.Errorf("Params: want 2, got %d", params)
	}
	if !bytes.Equal(readInstr, instr) {
		t.Errorf("Instr mismatch")
	}
}

func TestAllocClosure(t *testing.T) {
	InitCabinet()

	// Mock fnPtr
	fnPtr := Ptr(12345)

	// Mock freeVars
	freeVars := []Ptr{Ptr(10), Ptr(20), Ptr(30)}

	ptr, err := AllocClosure(fnPtr, freeVars)
	if err != nil {
		t.Fatalf("Alloc failed: %v", err)
	}

	readFn, readVars, err := ReadClosure(ptr)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if readFn != fnPtr {
		t.Errorf("FnPtr mismatch")
	}

	if len(readVars) != 3 {
		t.Fatalf("FreeVars len mismatch")
	}

	if readVars[0] != freeVars[0] || readVars[1] != freeVars[1] || readVars[2] != freeVars[2] {
		t.Errorf("FreeVars content mismatch")
	}
}
