package memory

import (
	"fmt"
	"unsafe"
)

// Resolve translates a Virtual Ptr to a Physical unsafe.Pointer.
// It handles Page Faults by bringing the required Drawer into RAM.
// It returns a pointer to the start of the data.
// WARNING: The returned pointer is valid ONLY until the next Alloc or Resolve call,
// because a subsequent call might trigger eviction/swapping of this drawer.
func (c *Cabinet) Resolve(p Ptr) (unsafe.Pointer, error) {
	if p == NilPtr {
		return nil, fmt.Errorf("segmentation fault: nil pointer")
	}

	id := p.DrawerID()
	offset := p.Offset()

	if id >= len(c.Drawers) {
		return nil, fmt.Errorf("segmentation fault: invalid drawer id %d", id)
	}

	// Lock the cabinet? (Single threaded for now)

	drawer := &c.Drawers[id]

	// Check if in RAM (PhysicalSlot != -1)
	if drawer.PhysicalSlot == -1 {
		// Page Fault! Bring to RAM.
		// This might evict someone else.
		if err := c.BringToRAM(id); err != nil {
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

	// Calculate Physical Address
	// RAM Base + (Slot * DrawerSize) + Offset
	base := uintptr(RAM.BasePointer())
	slotOffset := uintptr(slot) * uintptr(DRAWER_SIZE)
	physAddr := base + slotOffset + uintptr(offset)

	return unsafe.Pointer(physAddr), nil
}
