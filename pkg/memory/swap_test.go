package memory

import (
	"testing"
)

func TestSwapSpillAndRestore(t *testing.T) {
	InitCabinet()

	// Create data larger than one drawer?
	// No, let's just create enough drawers to force swap.
	// PHYSICAL_SLOTS = 16.
	// We allocate 17 drawers.

	// Fill RAM (16 drawers)
	for i := 0; i < PHYSICAL_SLOTS; i++ {
		// Each alloc triggers CreateDrawer because previous is full?
		// Alloc checks `activeTray.Remaining()`.
		// If we allocate DRAWER_SIZE/2 + 1, it fills Primary.
		// Wait, Alloc only checks current active tray.
		// We need to force `CreateDrawer`.
		// `Alloc` calls `CreateDrawer` if current is full.

		// Fill current drawer
		size := TRAY_SIZE
		_, err := Lemari.Alloc(size) // Fills Primary
		if err != nil {
			t.Fatalf("Alloc 1 failed: %v", err)
		}

		// Fill Secondary (Manual toggle? Alloc handles it? Alloc logic: `if activeDrawer.IsPrimaryActive`)
		// My Alloc logic doesn't toggle Primary/Secondary automatically yet.
		// It just calls CreateDrawer if current tray full.
		// So 1 alloc of TRAY_SIZE fills the drawer (effectively).
	}

	// Now we have 16 drawers full. RAM is full.
	// Allocate one more.
	// This should trigger `CreateDrawer` (Drawer 17).
	// `CreateDrawer` sees no free RAM slots.
	// It creates Drawer 17 in "Swapped" state (or just without slot).
	// `Alloc` calls `BringToRAM(17)`.
	// `BringToRAM` sees full RAM. Triggers `EvictToSwap(Slot 0 -> Drawer 0)`.
	// Drawer 0 goes to disk.
	// Drawer 17 takes Slot 0.

	ptr, err := Lemari.Alloc(100)
	if err != nil {
		t.Fatalf("Swap allocation failed: %v", err)
	}

	data := []byte("Hello Swap")
	Write(ptr, data)

	// Check Drawer 0 state
	if !Lemari.Drawers[0].IsSwapped {
		t.Error("Drawer 0 should be swapped out")
	}
	if Lemari.Drawers[0].SwapOffset == 0 {
		t.Error("Drawer 0 should have swap offset")
	}

	// Now access Drawer 0 again to trigger Restore
	// We need a way to force access. `BringToRAM`.
	err = Lemari.BringToRAM(0)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	if Lemari.Drawers[0].IsSwapped {
		t.Error("Drawer 0 should be in RAM")
	}

	// Drawer 17 should be swapped out (victim?)
	// Or Drawer 1 (Slot 0 was used by 17, then 0 came back).
	// `BringToRAM` evicts Slot 0. Slot 0 was holding Drawer 17.
	if !Lemari.Drawers[17].IsSwapped {
		t.Error("Drawer 17 should be swapped out")
	}

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
