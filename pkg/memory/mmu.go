package memory

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

var ErrPageFault = fmt.Errorf("page fault")

// resolve translates a Virtual Ptr to a Physical unsafe.Pointer.
// It handles Page Faults by bringing the required Drawer into RAM.
// It returns a pointer to the start of the data.
// WARNING: The returned pointer is valid ONLY while the Cabinet lock is held.
// Assumes c.mu is Locked.
func (c *Cabinet) resolve(p Ptr) (unsafe.Pointer, error) {
	if p == NilPtr {
		return nil, fmt.Errorf("segmentation fault: nil pointer")
	}

	id := p.DrawerID()
	offset := p.Offset()

	if id >= len(c.Drawers) {
		return nil, fmt.Errorf("segmentation fault: invalid drawer id %d", id)
	}

	drawer := &c.Drawers[id]

	// Check if in RAM (PhysicalSlot != -1)
	if drawer.PhysicalSlot == -1 {
		// Page Fault! Bring to RAM.
		// This might evict someone else.
		if err := c.bringToRAM(id); err != nil {
			return nil, fmt.Errorf("page fault error: %v", err)
		}
		// Re-fetch drawer pointer
		drawer = &c.Drawers[id]
	}

	// Now it is in RAM.
	slot := drawer.PhysicalSlot
	if slot < 0 || slot >= PHYSICAL_SLOTS {
		return nil, fmt.Errorf("MMU error: drawer marked as resident but has invalid slot %d", slot)
    }

	// GC: Increment Access Count (LFU)
	atomic.AddInt64(&drawer.AccessCount, 1)

	// Calculate Physical Address
	// RAM Base + (Slot * DrawerSize) + Offset
	offsetBytes := uintptr(slot)*uintptr(DRAWER_SIZE) + uintptr(offset)
	physAddr := unsafe.Add(RAM.BasePointer(), offsetBytes)

	return physAddr, nil
}

// resolveFast translates a Virtual Ptr to a Physical unsafe.Pointer WITHOUT bringing pages to RAM.
// Returns ErrPageFault if the page is not resident.
// Safe to call with RLock.
func (c *Cabinet) resolveFast(p Ptr) (unsafe.Pointer, error) {
	if p == NilPtr {
		return nil, fmt.Errorf("segmentation fault: nil pointer")
	}

	id := p.DrawerID()
	offset := p.Offset()

	if id >= len(c.Drawers) {
		return nil, fmt.Errorf("segmentation fault: invalid drawer id %d", id)
	}

	drawer := &c.Drawers[id]

	// Check if in RAM (PhysicalSlot != -1)
	if drawer.PhysicalSlot == -1 {
		return nil, ErrPageFault
	}

	// Now it is in RAM.
	slot := drawer.PhysicalSlot
	if slot < 0 || slot >= PHYSICAL_SLOTS {
		return nil, fmt.Errorf("MMU error: drawer marked as resident but has invalid slot %d", slot)
	}

	// GC: Increment Access Count (LFU)
	atomic.AddInt64(&drawer.AccessCount, 1)

	// Calculate Physical Address
	// RAM Base + (Slot * DrawerSize) + Offset
	offsetBytes := uintptr(slot)*uintptr(DRAWER_SIZE) + uintptr(offset)
	physAddr := unsafe.Add(RAM.BasePointer(), offsetBytes)

	return physAddr, nil
}

// resolveFast translates a Virtual Ptr to a Physical unsafe.Pointer WITHOUT bringing pages to RAM.
// Returns ErrPageFault if the page is not resident.
// Safe to call with RLock.
func (c *Cabinet) resolveFast(p Ptr) (unsafe.Pointer, error) {
	if p == NilPtr {
		return nil, fmt.Errorf("segmentation fault: nil pointer")
	}

	id := p.DrawerID()
	offset := p.Offset()

	if id >= len(c.Drawers) {
		return nil, fmt.Errorf("segmentation fault: invalid drawer id %d", id)
	}

	drawer := &c.Drawers[id]

	// Check if in RAM (PhysicalSlot != -1)
	if drawer.PhysicalSlot == -1 {
		return nil, ErrPageFault
	}

	// Now it is in RAM.
	slot := drawer.PhysicalSlot
	if slot < 0 || slot >= PHYSICAL_SLOTS {
		return nil, fmt.Errorf("MMU error: drawer marked as resident but has invalid slot %d", slot)
	}

	// Calculate Physical Address
	// RAM Base + (Slot * DrawerSize) + Offset
	offsetBytes := uintptr(slot)*uintptr(DRAWER_SIZE) + uintptr(offset)
	physAddr := unsafe.Add(RAM.BasePointer(), offsetBytes)

	return physAddr, nil
}
