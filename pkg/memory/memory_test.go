package memory

import (
	"bytes"
	"testing"
	"unsafe"
)

func TestCabinetInitialization(t *testing.T) {
	InitCabinet()

	if Lemari.ActiveDrawerIndex != 0 {
		t.Errorf("Expected active drawer 0, got %d", Lemari.ActiveDrawerIndex)
	}

	drawer := Lemari.Drawers[0]
	if !drawer.IsPrimaryActive {
		t.Error("Primary tray should be active initially")
	}

	// Drawer 0 reserves 8 bytes
	expectedSize := TRAY_SIZE - 8
	if drawer.PrimaryTray.Remaining() != expectedSize {
		t.Errorf("Expected tray size %d, got %d", expectedSize, drawer.PrimaryTray.Remaining())
	}
}

func TestAllocationAndWrite(t *testing.T) {
	InitCabinet()

	data := []byte("Hello Morph Memory")
	size := len(data)

	ptr, err := Lemari.Alloc(size)
	if err != nil {
		t.Fatalf("Allocation failed: %v", err)
	}

	if ptr == NilPtr {
		t.Fatal("Allocated pointer should not be nil")
	}

	err = Write(ptr, data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify data manually
	Lemari.mu.Lock()
	rawPtr, err := Lemari.resolve(ptr)
	if err != nil {
		Lemari.mu.Unlock()
		t.Fatal(err)
	}
	// Copy data out immediately to be safe
	srcSlice := unsafe.Slice((*byte)(rawPtr), size)
	readData := make([]byte, size)
	copy(readData, srcSlice)
	Lemari.mu.Unlock()

	if !bytes.Equal(readData, data) {
		t.Errorf("Memory corruption! Expected %s, got %s", data, readData)
	}
}

func TestManualCopy(t *testing.T) {
	InitCabinet()

	// 1. Allocate and Write "Glass A"
	dataA := []byte("Gelas A")
	ptrA, _ := Lemari.Alloc(len(dataA))
	Write(ptrA, dataA)

	// 2. Allocate space for "Glass B"
	ptrB, _ := Lemari.Alloc(len(dataA))

	// 3. Manual Copy (Move A to B)
	err := MemCpy(ptrA, ptrB, len(dataA))
	if err != nil {
		t.Fatalf("MemCpy failed: %v", err)
	}

	// 4. Verify B has A's content
	Lemari.mu.Lock()
	rawB, err := Lemari.resolve(ptrB)
	if err != nil {
		Lemari.mu.Unlock()
		t.Fatal(err)
	}
	srcSlice := unsafe.Slice((*byte)(rawB), len(dataA))
	readB := make([]byte, len(dataA))
	copy(readB, srcSlice)
	Lemari.mu.Unlock()

	if string(readB) != "Gelas A" {
		t.Errorf("Copy failed. Expected 'Gelas A', got '%s'", readB)
	}
}

func TestCabinetFull(t *testing.T) {
	InitCabinet()

	// Try to allocate more than a tray
	largeSize := TRAY_SIZE + 100
	_, err := Lemari.Alloc(largeSize)

	if err == nil {
		t.Error("Should fail when tray is full")
	}
}
