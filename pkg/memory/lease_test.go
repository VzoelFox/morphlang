package memory

import (
	"testing"
	"unsafe"
)

func TestDrawerLeaseRollback(t *testing.T) {
	InitCabinet()

	// 1. Acquire
	unitID := 1
	lease, err := Lemari.AcquireDrawer(unitID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	if lease.UnitID != unitID {
		t.Errorf("Lease UnitID mismatch")
	}
	if !lease.IsActive {
		t.Errorf("Lease should be active")
	}

	drawerID := lease.DrawerID
	drawer := &Lemari.Drawers[drawerID]

	if drawer.Lease != lease {
		t.Errorf("Drawer lease not set correctly")
	}

	// 2. Modify Data (Simulate Work)
	if drawer.PhysicalSlot == -1 {
		t.Fatal("Drawer should be in RAM")
	}

	ramPtr := unsafe.Add(RAM.BasePointer(), uintptr(drawer.PhysicalSlot)*uintptr(DRAWER_SIZE))
	dataSlice := unsafe.Slice((*byte)(ramPtr), DRAWER_SIZE)

	// Write a marker (Simulate dirty state)
	dataSlice[0] = 0xAA

	// 3. Rollback
	// Should revert to state AT ACQUISITION (which was 0x00)
	err = Lemari.RollbackDrawer(lease)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify Data is 0x00 again
	if dataSlice[0] != 0x00 {
		t.Errorf("Rollback failed. Expected 0x00, got 0x%X", dataSlice[0])
	}

	if lease.IsActive {
		t.Errorf("Lease should be inactive after rollback")
	}
	if drawer.Lease != nil {
		t.Errorf("Drawer should be free")
	}
}

func TestDrawerLeaseCommit(t *testing.T) {
	InitCabinet()
	unitID := 2
	lease, err := Lemari.AcquireDrawer(unitID)
	if err != nil {
		t.Fatalf("Acquire failed: %v", err)
	}

	// Modify
	drawer := &Lemari.Drawers[lease.DrawerID]
	ramPtr := unsafe.Add(RAM.BasePointer(), uintptr(drawer.PhysicalSlot)*uintptr(DRAWER_SIZE))
	dataSlice := unsafe.Slice((*byte)(ramPtr), DRAWER_SIZE)
	dataSlice[0] = 0xBB

	// Commit
	err = Lemari.CommitDrawer(lease)
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Verify Data PERSISTS
	if dataSlice[0] != 0xBB {
		t.Errorf("Commit failed to persist data. Got 0x%X", dataSlice[0])
	}
	if lease.IsActive {
		t.Errorf("Lease should be inactive")
	}
	if drawer.Lease != nil {
		t.Errorf("Drawer should be free")
	}
}
