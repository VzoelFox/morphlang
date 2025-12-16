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
	header := GetHeader(ptr)
	if header.Type != TagInteger {
		t.Errorf("Wrong tag. Expected %d, got %d", TagInteger, header.Type)
	}
	// Size should be HeaderSize (padded?) + 8
	// Header struct has Type(1) + Size(4) = 5 bytes?
	// Go struct padding implies Header might be 8 bytes?
	// unsafe.Sizeof(Header{}) usually 8 if aligned.
	if int(header.Size) < 8+8 {
		t.Errorf("Suspicious size: %d", header.Size)
	}

	// 3. Read Value
	readVal := ReadInteger(ptr)
	if readVal != val {
		t.Errorf("Data corruption! Expected %d, got %d", val, readVal)
	}

	// 4. Test another allocation (Bump pointer works?)
	val2 := int64(999)
	ptr2, _ := AllocInteger(val2)
	if ptr2 == ptr {
		t.Error("Allocator returned same address for second object")
	}

	readVal2 := ReadInteger(ptr2)
	if readVal2 != val2 {
		t.Errorf("Second object corruption. Expected %d, got %d", val2, readVal2)
	}

	// Verify first object still intact
	if ReadInteger(ptr) != val {
		t.Error("First object overwritten!")
	}
}
