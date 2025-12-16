package memory

import (
	"fmt"
	"unsafe"
)

// Alloc allocates memory. It handles "Draft Otomatis" (Swapping) if RAM is full.
func (c *Cabinet) Alloc(size int) (Ptr, error) {
	if size <= 0 {
		return NilPtr, fmt.Errorf("invalid allocation size: %d", size)
	}

	alignedSize := (size + 7) &^ 7

	if alignedSize > TRAY_SIZE {
		return NilPtr, fmt.Errorf("allocation size %d exceeds tray limit %d", alignedSize, TRAY_SIZE)
	}

	activeDrawer := &c.Drawers[c.ActiveDrawerIndex]

	// Ensure active drawer is in RAM (Resident)
	if activeDrawer.PhysicalSlot == -1 {
		if err := c.BringToRAM(c.ActiveDrawerIndex); err != nil {
			return NilPtr, err
		}
		// Re-fetch pointer
		activeDrawer = &c.Drawers[c.ActiveDrawerIndex]
	}

	var activeTray *Tray
	if activeDrawer.IsPrimaryActive {
		activeTray = &activeDrawer.PrimaryTray
	} else {
		activeTray = &activeDrawer.SecondaryTray
	}

	if activeTray.Remaining() < alignedSize {
		// Drawer Full. Create new drawer.
		newDrawer := CreateDrawer()
		if newDrawer == nil {
			return NilPtr, fmt.Errorf("virtual memory limit reached")
		}

		c.ActiveDrawerIndex = newDrawer.ID

		// Ensure new drawer is resident (CreateDrawer tries to put it in RAM, but might fail/swap immediately if logic changes)
		if newDrawer.PhysicalSlot == -1 {
			if err := c.BringToRAM(newDrawer.ID); err != nil {
				return NilPtr, err
			}
			newDrawer = &c.Drawers[c.ActiveDrawerIndex]
		}

		// Recurse
		return c.Alloc(size)
	}

	ptr := activeTray.Current
	// Bump pointer (Valid because offset is in lower bits)
	activeTray.Current += Ptr(alignedSize)

	return ptr, nil
}

// BringToRAM ensures the drawer is in physical RAM.
// If RAM is full, it evicts a victim drawer to swap.
func (c *Cabinet) BringToRAM(drawerID int) error {
	target := &c.Drawers[drawerID]
	if target.PhysicalSlot != -1 {
		return nil // Already resident
	}

	// Find free slot
	freeSlot := -1
	for i := 0; i < PHYSICAL_SLOTS; i++ {
		if c.RAMSlots[i] == -1 {
			freeSlot = i
			break
		}
	}

	// If no free slot, EVICT victim (Draft Otomatis)
	if freeSlot == -1 {
		// Simple policy: Evict slot 0 (FIFO replacement for simplicity)
		// Better policy: Evict LRU? For now, slot 0 is fine.
		victimID := c.RAMSlots[0]
		if victimID == drawerID {
			return fmt.Errorf("deadlock: trying to evict self") // Should not happen
		}

		err := c.EvictToSwap(victimID)
		if err != nil {
			return err
		}
		freeSlot = 0 // Slot 0 is now free
	}

	// Calculate Physical Address for this slot
	base := uintptr(RAM.BasePointer())
	offset := uintptr(freeSlot) * uintptr(DRAWER_SIZE)
	physAddr := unsafe.Pointer(base + offset)
	destSlice := unsafe.Slice((*byte)(physAddr), DRAWER_SIZE)

	// Load data if it was previously saved
	if target.SwapOffset > 0 {
		err := Swap.Restore(target.SwapOffset, destSlice)
		if err != nil {
			return err
		}
	} else {
		// Fresh drawer (or never swapped out with data).
		// Zero it out? Or explicitly assume CreateDrawer initialized it?
		// CreateDrawer doesn't touch RAM if it made a swapped drawer.
		// So we should zero it out to be safe.
		for i := range destSlice {
			destSlice[i] = 0
		}
	}

	c.RAMSlots[freeSlot] = drawerID
	target.PhysicalSlot = freeSlot
	// target.IsSwapped was used as "Is currently swapped out".
	// The struct comment said "IsSwapped bool". Let's use PhysicalSlot != -1 as the truth.
	// But `IsSwapped` might mean "Has backup on disk".
	// Let's interpret `IsSwapped` as "Currently NOT in RAM".
	target.IsSwapped = false

	return nil
}

// EvictToSwap moves a drawer from RAM to Disk
func (c *Cabinet) EvictToSwap(drawerID int) error {
	victim := &c.Drawers[drawerID]
	slot := victim.PhysicalSlot

	if slot == -1 {
		return nil // Already swapped
	}

	// 1. Get RAM content
	base := uintptr(RAM.BasePointer())
	offset := uintptr(slot) * uintptr(DRAWER_SIZE)
	physAddr := unsafe.Pointer(base + offset)
	srcSlice := unsafe.Slice((*byte)(physAddr), DRAWER_SIZE)

	// 2. Write to .z file
	// Note: Swap.Spill typically Appends. If we want to Overwrite old swap location,
	// we need Swap.WriteAt(offset).
	// Assuming Swap.Spill always allocates NEW space on disk?
	// If we keep appending, disk usage grows indefinitely.
	// For "Snapshot/Rewind", append is good!
	// For paging, we want to reuse.
	// Let's assume Swap.Spill returns a new offset.
	fileOffset, err := Swap.Spill(srcSlice)
	if err != nil {
		return err
	}

	// 3. Update State
	victim.IsSwapped = true
	victim.SwapOffset = fileOffset
	victim.PhysicalSlot = -1

	// 4. Free RAM slot
	c.RAMSlots[slot] = -1

	// Note: Virtual Pointers remain valid! They just point to this DrawerID.
	// Next access will trigger BringToRAM.

	return nil
}

// Write writes data to the pointer address.
func Write(dst Ptr, data []byte) error {
	// Translate Virtual Ptr to Physical Address (Triggering Page Fault if needed)
	rawPtr, err := Lemari.Resolve(dst)
	if err != nil {
		return err
	}

	size := len(data)
	targetSlice := unsafe.Slice((*byte)(rawPtr), size)
	copy(targetSlice, data)
	return nil
}

// Read reads data from the pointer address.
func Read(src Ptr, size int) ([]byte, error) {
	rawPtr, err := Lemari.Resolve(src)
	if err != nil {
		return nil, err
	}

	srcSlice := unsafe.Slice((*byte)(rawPtr), size)
	// Return a copy so user owns the bytes (and doesn't hold unsafe ptr to potentially moved RAM)
	ret := make([]byte, size)
	copy(ret, srcSlice)
	return ret, nil
}
