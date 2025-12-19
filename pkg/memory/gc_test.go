package memory

import (
	"testing"
)

func TestGC_LFU_Eviction(t *testing.T) {
	// 1. Reset Cabinet
	InitCabinet() // Creates 16 drawers (0-15) filling RAM

	// 2. Setup Access Patterns
	// D0: High Access (10)
	for i := 0; i < 10; i++ {
		ptr := NewPtr(0, 8) // Offset 8 to avoid NilPtr check
		Lemari.mu.RLock()
		Lemari.resolveFast(ptr)
		Lemari.mu.RUnlock()
	}

	// D1: Victim Candidate (1 Access)
	ptr1 := NewPtr(1, 8)
	Lemari.mu.RLock()
	Lemari.resolveFast(ptr1)
	Lemari.mu.RUnlock()

	// D2-15: Medium Access (5 Accesses)
	for id := 2; id < 16; id++ {
		for k := 0; k < 5; k++ {
			p := NewPtr(id, 8)
			Lemari.mu.RLock()
			Lemari.resolveFast(p)
			Lemari.mu.RUnlock()
		}
	}

	// 3. Create Overflow Drawer (D16)
	// This will be created in "Swapped" state because RAM is full
	d16 := CreateDrawer()
	if d16.PhysicalSlot != -1 {
		t.Fatalf("Expected D16 to be created in Swap (Virtual), got slot %d", d16.PhysicalSlot)
	}

	// 4. Force Swap-In of D16
	// This should trigger eviction of the LFU victim (D1)
	Lemari.mu.Lock()
	err := Lemari.bringToRAM(16)
	Lemari.mu.Unlock()

	if err != nil {
		t.Fatalf("bringToRAM failed: %v", err)
	}

	// 5. Verify Results
	// D1 should be evicted (Swapped Out)
	d1 := &Lemari.Drawers[1]
	if d1.PhysicalSlot != -1 {
		t.Errorf("LFU Failure: D1 (Score 1) should be evicted. Status: Resident in slot %d", d1.PhysicalSlot)
	}

	// D0 should be safe (Score 10)
	d0 := &Lemari.Drawers[0]
	if d0.PhysicalSlot == -1 {
		t.Errorf("LFU Failure: D0 (Score 10) was evicted incorrectly.")
	}

	// D16 should be Resident
	if d16.PhysicalSlot == -1 {
		t.Errorf("D16 should be resident after bringToRAM.")
	}
}

func TestGC_Daemon_Aging(t *testing.T) {
	InitCabinet()

	// Set artificial high score
	Lemari.Drawers[0].AccessCount = 100

	// Manually trigger GC cycle (simulate Daemon wake up)
	Lemari.LFUAging()

	// Score should halve
	if Lemari.Drawers[0].AccessCount != 50 {
		t.Errorf("Aging Failure: Expected 50, got %d", Lemari.Drawers[0].AccessCount)
	}

	Lemari.LFUAging()
	if Lemari.Drawers[0].AccessCount != 25 {
		t.Errorf("Aging Failure: Expected 25, got %d", Lemari.Drawers[0].AccessCount)
	}
}
