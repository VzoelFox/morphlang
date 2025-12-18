package memory

import (
	"testing"
)

func TestStringAllocation(t *testing.T) {
	InitCabinet()

	input := "Hello, World!"
	ptr, err := AllocString(input)
	if err != nil {
		t.Fatalf("AllocString failed: %v", err)
	}

	readBack, err := ReadString(ptr)
	if err != nil {
		t.Fatalf("ReadString failed: %v", err)
	}

	if readBack != input {
		t.Errorf("expected %q, got %q", input, readBack)
	}
}

func TestBooleanAllocation(t *testing.T) {
	InitCabinet()

	ptrTrue, err := AllocBoolean(true)
	if err != nil {
		t.Fatalf("AllocBoolean failed: %v", err)
	}
	valTrue, err := ReadBoolean(ptrTrue)
	if err != nil {
		t.Fatalf("ReadBoolean failed: %v", err)
	}
	if !valTrue {
		t.Errorf("expected true")
	}

	ptrFalse, err := AllocBoolean(false)
	if err != nil {
		t.Fatalf("AllocBoolean failed: %v", err)
	}
	valFalse, err := ReadBoolean(ptrFalse)
	if err != nil {
		t.Fatalf("ReadBoolean failed: %v", err)
	}
	if valFalse {
		t.Errorf("expected false")
	}
}

func TestArrayAllocation(t *testing.T) {
	InitCabinet()

	length := 5
	ptrArr, err := AllocArray(length, length)
	if err != nil {
		t.Fatalf("AllocArray failed: %v", err)
	}

	// Create some integers
	ptr1, _ := AllocInteger(10)
	ptr2, _ := AllocInteger(20)

	// Write to array
	err = WriteArrayElement(ptrArr, 0, ptr1)
	if err != nil {
		t.Fatalf("WriteArrayElement failed: %v", err)
	}
	err = WriteArrayElement(ptrArr, 4, ptr2)
	if err != nil {
		t.Fatalf("WriteArrayElement failed: %v", err)
	}

	// Read back
	p1, err := ReadArrayElement(ptrArr, 0)
	if err != nil {
		t.Fatalf("ReadArrayElement failed: %v", err)
	}
	if p1 != ptr1 {
		t.Errorf("expected ptr1 %v, got %v", ptr1, p1)
	}

	p2, err := ReadArrayElement(ptrArr, 4)
	if err != nil {
		t.Fatalf("ReadArrayElement failed: %v", err)
	}
	if p2 != ptr2 {
		t.Errorf("expected ptr2 %v, got %v", ptr2, p2)
	}

	pEmpty, err := ReadArrayElement(ptrArr, 2)
	if err != nil {
		t.Fatalf("ReadArrayElement failed: %v", err)
	}
	if pEmpty != NilPtr {
		t.Errorf("expected NilPtr, got %v", pEmpty)
	}

	// Read Integer value from array element
	val1, _ := ReadInteger(p1)
	if val1 != 10 {
		t.Errorf("expected 10, got %d", val1)
	}
}
