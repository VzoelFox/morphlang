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

	if drawer.PrimaryTray.Remaining() != TRAY_SIZE {
		t.Errorf("Expected tray size %d, got %d", TRAY_SIZE, drawer.PrimaryTray.Remaining())
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

	// Verify data
	rawPtr := ptr.ToUnsafe()
	readData := unsafe.Slice((*byte)(rawPtr), size)

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

	// 2. Allocate space for "Glass B" (in same tray for now)
	ptrB, _ := Lemari.Alloc(len(dataA))

	// 3. Manual Copy (Move A to B)
	err := MemCpy(ptrA, ptrB, len(dataA))
	if err != nil {
		t.Fatalf("MemCpy failed: %v", err)
	}

	// 4. Verify B has A's content
	rawB := ptrB.ToUnsafe()
	readB := unsafe.Slice((*byte)(rawB), len(dataA))
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
