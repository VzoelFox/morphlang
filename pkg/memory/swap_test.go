package memory

import (
	"testing"
)

func TestSwapSpillAndRestore(t *testing.T) {
	InitCabinet()

	// Fill RAM (16 drawers)
	for i := 0; i < PHYSICAL_SLOTS; i++ {
		// Fill current drawer
		size := TRAY_SIZE
		_, err := Lemari.Alloc(size)
		if err != nil {
			t.Fatalf("Alloc loop failed at %d: %v", i, err)
		}
	}

	// Now we have 16 drawers full. RAM is full.
	// Allocate one more.
	// This should trigger `CreateDrawer` (Drawer 16).
	// `Alloc` calls `bringToRAM(16)`.
	// `bringToRAM` sees full RAM. Triggers `evictToSwap(Slot 0 -> Drawer 0)`.
	// Drawer 0 goes to disk.
	// Drawer 16 takes Slot 0.

	ptr, err := Lemari.Alloc(100)
	if err != nil {
		t.Fatalf("Swap allocation failed: %v", err)
	}

	data := []byte("Hello Swap")
	Write(ptr, data)

	// Check Drawer 0 state
	Lemari.mu.Lock()
	if !Lemari.Drawers[0].IsSwapped {
		t.Error("Drawer 0 should be swapped out")
	}
	if Lemari.Drawers[0].SwapOffset == 0 {
		t.Error("Drawer 0 should have swap offset")
	}
	Lemari.mu.Unlock()

	// Now access Drawer 0 again to trigger Restore
	// We use internal bringToRAM
	Lemari.mu.Lock()
	err = Lemari.bringToRAM(0)
	Lemari.mu.Unlock()

	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	Lemari.mu.Lock()
	if Lemari.Drawers[0].IsSwapped {
		t.Error("Drawer 0 should be in RAM")
	}

	// Drawer 16 should be swapped out (victim?)
	// `bringToRAM` evicts Slot 0. Slot 0 was holding Drawer 16.
	// DrawerID is 16 (length was 16 before alloc, so new index is 16).
	if !Lemari.Drawers[16].IsSwapped {
		t.Error("Drawer 16 should be swapped out")
	}
	Lemari.mu.Unlock()

	// Cleanup
	Swap.FreeCache()
}

func TestSwapFileStructure(t *testing.T) {
	InitCabinet()
	Swap.FreeCache() // clear previous
	InitCabinet() // re-init

	// Force spill
	// Fill RAM
	for i := 0; i < PHYSICAL_SLOTS; i++ {
		Lemari.Alloc(TRAY_SIZE)
	}
	// Trigger spill
	Lemari.Alloc(100)

	// Check file existence
	if _, err := Swap.file.Stat(); err != nil {
		t.Error("Swap file not created")
	}

	Swap.FreeCache()
}
