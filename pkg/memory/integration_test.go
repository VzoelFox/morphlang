package memory

import (
	"testing"
)

func TestIntegerLifecycle(t *testing.T) {
	InitCabinet()

	val := int64(123456789)

	// 1. Allocate
	ptr, err := AllocInteger(val)
	if err != nil {
		t.Fatalf("AllocInteger failed: %v", err)
	}
	if ptr == NilPtr {
		t.Fatal("Allocated pointer is nil")
	}

	// 2. Read Header
	header, err := ReadHeader(ptr)
	if err != nil {
		t.Fatalf("ReadHeader failed: %v", err)
	}
	if header.Type != TagInteger {
		t.Errorf("Wrong tag. Expected %d, got %d", TagInteger, header.Type)
	}
	// Size should be HeaderSize (padded?) + 8
	if int(header.Size) < 8+8 {
		t.Errorf("Suspicious size: %d", header.Size)
	}

	// 3. Read Value
	readVal, err := ReadInteger(ptr)
	if err != nil {
		t.Fatalf("ReadInteger failed: %v", err)
	}
	if readVal != val {
		t.Errorf("Data corruption! Expected %d, got %d", val, readVal)
	}

	// 4. Test another allocation (Bump pointer works?)
	val2 := int64(999)
	ptr2, _ := AllocInteger(val2)
	if ptr2 == ptr {
		t.Error("Allocator returned same address for second object")
	}

	readVal2, err := ReadInteger(ptr2)
	if err != nil {
		t.Fatalf("ReadInteger 2 failed: %v", err)
	}
	if readVal2 != val2 {
		t.Errorf("Second object corruption. Expected %d, got %d", val2, readVal2)
	}

	// Verify first object still intact
	val1, err := ReadInteger(ptr)
	if err != nil {
		t.Fatalf("ReadInteger 1 failed: %v", err)
	}
	if val1 != val {
		t.Error("First object overwritten!")
	}
}
